package repository

import (
	"errors"
	"time"

	"awesomeProject/internal/app/ds"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("запись не найдена")
	ErrAlreadyExists = errors.New("запись уже существует")
	ErrForbidden     = errors.New("нет прав на операцию")
)

// publishedWithLikes — опубликованные услуги вместе с лайками
func (r *Repository) publishedWithLikes() *gorm.DB {
	return r.db.Preload("Likes").Where("status = ?", ds.StatusPublished)
}

// GetDignitarieLists — опубликованные услуги; при onDate != nil только те,
// кто находился в должности в эту дату
func (r *Repository) GetDignitarieLists(onDate *time.Time) ([]ds.DignitarieList, error) {
	query := r.publishedWithLikes()
	if onDate != nil {
		query = query.Where("start_date <= ? AND end_date >= ?", *onDate, *onDate)
	}

	var dignitarieLists []ds.DignitarieList
	if err := query.Order("id").Find(&dignitarieLists).Error; err != nil {
		return nil, err
	}
	return dignitarieLists, nil
}

// GetDignitarieList — опубликованная услуга по ID
func (r *Repository) GetDignitarieList(id uint) (ds.DignitarieList, error) {
	var dignitarieList ds.DignitarieList
	err := r.publishedWithLikes().Where("id = ?", id).First(&dignitarieList).Error
	return dignitarieList, wrapNotFound(err)
}

// GetNextDignitarieList — следующая опубликованная услуга после ID, после последней — снова первая
func (r *Repository) GetNextDignitarieList(id uint) (ds.DignitarieList, error) {
	var dignitarieList ds.DignitarieList
	err := r.publishedWithLikes().Where("id > ?", id).Order("id").First(&dignitarieList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) && id > 0 {
		return r.GetNextDignitarieList(0)
	}
	return dignitarieList, wrapNotFound(err)
}

// GetDraftDignitarieList — черновик пользователя (не более одного)
func (r *Repository) GetDraftDignitarieList(creatorID uint) (ds.DignitarieList, error) {
	var dignitarieList ds.DignitarieList
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&dignitarieList).Error
	return dignitarieList, wrapNotFound(err)
}

// CreateDraftDignitarieList — создание черновика с url загруженных фото и видео
func (r *Repository) CreateDraftDignitarieList(dignitarieList *ds.DignitarieList) error {
	if _, err := r.GetDraftDignitarieList(dignitarieList.CreatorID); err == nil {
		return ErrAlreadyExists
	}
	dignitarieList.Status = ds.StatusDraft
	err := r.db.Create(dignitarieList).Error
	if isUniqueViolation(err) { // параллельный запрос успел создать черновик
		return ErrAlreadyExists
	}
	return err
}

// PublishDignitarieList — заполнение необязательных полей черновика и смена статуса на «опубликован»
func (r *Repository) PublishDignitarieList(creatorID uint, description string, startDate, endDate time.Time) (ds.DignitarieList, error) {
	draft, err := r.GetDraftDignitarieList(creatorID)
	if err != nil {
		return ds.DignitarieList{}, err
	}

	now := time.Now()
	err = r.db.Model(&draft).Updates(map[string]any{
		"description":  description,
		"start_date":   startDate,
		"end_date":     endDate,
		"status":       ds.StatusPublished,
		"published_at": &now,
	}).Error
	if err != nil {
		return ds.DignitarieList{}, err
	}
	return r.GetDignitarieList(draft.ID)
}

// DeleteDignitarieList — логическое удаление: только свои услуги в статусе черновик или опубликован
func (r *Repository) DeleteDignitarieList(id, userID uint) error {
	var dignitarieList ds.DignitarieList
	err := r.db.Where("id = ? AND status <> ?", id, ds.StatusDeleted).First(&dignitarieList).Error
	if err != nil {
		return wrapNotFound(err)
	}
	if dignitarieList.CreatorID != userID {
		return ErrForbidden
	}
	return r.db.Model(&dignitarieList).Update("status", ds.StatusDeleted).Error
}

// SetLike ставит (liked = true) или снимает лайк пользователя у опубликованной услуги
func (r *Repository) SetLike(dignitarieListID, userID uint, liked bool) (ds.DignitarieList, error) {
	if _, err := r.GetDignitarieList(dignitarieListID); err != nil {
		return ds.DignitarieList{}, err
	}

	like := ds.DignitarieListLike{UserID: userID, DignitarieListID: dignitarieListID}
	var err error
	if liked {
		err = r.db.Where(&like).FirstOrCreate(&like).Error
	} else {
		err = r.db.Where("user_id = ? AND dignitarie_list_id = ?", userID, dignitarieListID).Delete(&ds.DignitarieListLike{}).Error
	}
	if err != nil {
		return ds.DignitarieList{}, err
	}
	return r.GetDignitarieList(dignitarieListID)
}

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

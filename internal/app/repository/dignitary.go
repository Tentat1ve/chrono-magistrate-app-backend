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

// publishedWithLikes — опубликованные сановники вместе с лайками
func (r *Repository) publishedWithLikes() *gorm.DB {
	return r.db.Preload("Likes").Where("status = ?", ds.DignitaryStatusPublished)
}

// GetDignitaries — опубликованные сановники; фильтр по полям по теме: год пребывания в должности
func (r *Repository) GetDignitaries(year *int) ([]ds.Dignitary, error) {
	query := r.publishedWithLikes()
	if year != nil {
		query = query.Where("office_start <= ? AND office_end >= ?", *year, *year)
	}

	var dignitaries []ds.Dignitary
	if err := query.Order("id").Find(&dignitaries).Error; err != nil {
		return nil, err
	}
	return dignitaries, nil
}

// GetDignitary — опубликованный сановник по ID
func (r *Repository) GetDignitary(id uint) (ds.Dignitary, error) {
	var dignitary ds.Dignitary
	err := r.publishedWithLikes().Where("id = ?", id).First(&dignitary).Error
	return dignitary, wrapNotFound(err)
}

// GetNextDignitary — следующий опубликованный сановник после ID, после последнего — снова первый
func (r *Repository) GetNextDignitary(id uint) (ds.Dignitary, error) {
	var dignitary ds.Dignitary
	err := r.publishedWithLikes().Where("id > ?", id).Order("id").First(&dignitary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) && id > 0 {
		return r.GetNextDignitary(0)
	}
	return dignitary, wrapNotFound(err)
}

// GetDraftDignitary — черновик пользователя (не более одного)
func (r *Repository) GetDraftDignitary(creatorID uint) (ds.Dignitary, error) {
	var dignitary ds.Dignitary
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.DignitaryStatusDraft).First(&dignitary).Error
	return dignitary, wrapNotFound(err)
}

// CreateDraftDignitary — создание черновика с url загруженных фото и видео
func (r *Repository) CreateDraftDignitary(dignitary *ds.Dignitary) error {
	if _, err := r.GetDraftDignitary(dignitary.CreatorID); err == nil {
		return ErrAlreadyExists
	}
	dignitary.Status = ds.DignitaryStatusDraft
	err := r.db.Create(dignitary).Error
	if isUniqueViolation(err) { // параллельный запрос успел создать черновик
		return ErrAlreadyExists
	}
	return err
}

// PublishDignitary — заполнение полей черновика и смена статуса на «опубликован»
func (r *Repository) PublishDignitary(creatorID uint, office, description string, officeStart, officeEnd int) (ds.Dignitary, error) {
	draft, err := r.GetDraftDignitary(creatorID)
	if err != nil {
		return ds.Dignitary{}, err
	}

	now := time.Now()
	err = r.db.Model(&draft).Updates(map[string]any{
		"office":       office,
		"description":  description,
		"office_start": officeStart,
		"office_end":   officeEnd,
		"status":       ds.DignitaryStatusPublished,
		"published_at": &now,
	}).Error
	if err != nil {
		return ds.Dignitary{}, err
	}
	return r.GetDignitary(draft.ID)
}

// DeleteDignitary — логическое удаление: только свои сановники в статусе черновик или опубликован
func (r *Repository) DeleteDignitary(id, userID uint) error {
	var dignitary ds.Dignitary
	err := r.db.Where("id = ? AND status <> ?", id, ds.DignitaryStatusDeleted).First(&dignitary).Error
	if err != nil {
		return wrapNotFound(err)
	}
	if dignitary.CreatorID != userID {
		return ErrForbidden
	}
	return r.db.Model(&dignitary).Update("status", ds.DignitaryStatusDeleted).Error
}

// SetLike ставит (liked = true) или снимает лайк пользователя у опубликованного сановника
func (r *Repository) SetLike(dignitaryID, userID uint, liked bool) (ds.Dignitary, error) {
	if _, err := r.GetDignitary(dignitaryID); err != nil {
		return ds.Dignitary{}, err
	}

	like := ds.DignitaryLike{UserID: userID, DignitaryID: dignitaryID}
	var err error
	if liked {
		err = r.db.Where(&like).FirstOrCreate(&like).Error
	} else {
		err = r.db.Where("user_id = ? AND dignitary_id = ?", userID, dignitaryID).Delete(&ds.DignitaryLike{}).Error
	}
	if err != nil {
		return ds.Dignitary{}, err
	}
	return r.GetDignitary(dignitaryID)
}

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

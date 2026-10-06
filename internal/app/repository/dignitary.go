package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"awesomeProject/internal/app/ds"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("запись не найдена")

// publishedWithLikes — опубликованные сановники вместе с лайками
func (r *Repository) publishedWithLikes() *gorm.DB {
	return r.db.Preload("Likes").Where("status = ?", ds.DignitaryStatusPublished)
}

// GetDignitaries — опубликованные сановники; при year != nil только те,
// кто находился в должности в этом году (ORM)
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

// GetDignitary — опубликованный сановник по ID, из БД возвращается одна строка (ORM)
func (r *Repository) GetDignitary(id uint) (ds.Dignitary, error) {
	var dignitary ds.Dignitary
	err := r.publishedWithLikes().Where("id = ?", id).First(&dignitary).Error
	return dignitary, wrapNotFound(err)
}

// GetNextDignitary — следующий опубликованный сановник после ID,
// после последнего — снова первый (ORM, одна строка из БД)
func (r *Repository) GetNextDignitary(id uint) (ds.Dignitary, error) {
	var dignitary ds.Dignitary
	err := r.publishedWithLikes().Where("id > ?", id).Order("id").First(&dignitary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) && id > 0 {
		return r.GetNextDignitary(0)
	}
	return dignitary, wrapNotFound(err)
}

// GetDraftDignitary — черновик пользователя (не более одного) (ORM)
func (r *Repository) GetDraftDignitary(creatorID uint) (ds.Dignitary, error) {
	var dignitary ds.Dignitary
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.DignitaryStatusDraft).First(&dignitary).Error
	return dignitary, wrapNotFound(err)
}

// CreateDraftDignitary — создание черновика: название, url фото и видео (ORM)
func (r *Repository) CreateDraftDignitary(creatorID uint, name, imageURL, videoURL string) error {
	dignitary := ds.Dignitary{
		Name:      name,
		ImageURL:  imageURL,
		VideoURL:  videoURL,
		Status:    ds.DignitaryStatusDraft,
		CreatorID: creatorID,
	}
	return r.db.Create(&dignitary).Error
}

// PublishDignitary — заполнение полей черновика и смена статуса на «опубликован» (ORM)
func (r *Repository) PublishDignitary(creatorID uint, office, description string, officeStart, officeEnd int) error {
	draft, err := r.GetDraftDignitary(creatorID)
	if err != nil {
		return err
	}

	now := time.Now()
	return r.db.Model(&draft).Updates(map[string]any{
		"office":       office,
		"description":  description,
		"office_start": officeStart,
		"office_end":   officeEnd,
		"status":       ds.DignitaryStatusPublished,
		"published_at": &now,
	}).Error
}

// DeleteDignitary — логическое удаление сановника SQL-запросом UPDATE без ORM (через курсор)
func (r *Repository) DeleteDignitary(id uint) error {
	query := "UPDATE dignitaries SET status = $1 WHERE id = $2 AND status = $3 RETURNING id"

	// курсор на строку результата
	row := r.db.Raw(query, ds.DignitaryStatusDeleted, id, ds.DignitaryStatusPublished).Row()

	var deletedID uint
	if err := row.Scan(&deletedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("сановник id=%d: %w", id, ErrNotFound)
		}
		return err
	}
	return nil
}

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

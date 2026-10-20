package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"awesomeProject/internal/app/ds"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("запись не найдена")

// publishedWithLikes — опубликованные услуги вместе с лайками
func (r *Repository) publishedWithLikes() *gorm.DB {
	return r.db.Preload("Likes").Where("status = ?", ds.StatusPublished)
}

// GetDignitarieLists — опубликованные услуги; при onDate != nil только те,
// кто находился в должности в эту дату (ORM)
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

// GetDignitarieList — опубликованная услуга по ID, из БД возвращается одна строка (ORM)
func (r *Repository) GetDignitarieList(id uint) (ds.DignitarieList, error) {
	var dignitarieList ds.DignitarieList
	err := r.publishedWithLikes().Where("id = ?", id).First(&dignitarieList).Error
	return dignitarieList, wrapNotFound(err)
}

// GetNextDignitarieList — следующая опубликованная услуга после ID,
// после последней — снова первая (ORM, одна строка из БД)
func (r *Repository) GetNextDignitarieList(id uint) (ds.DignitarieList, error) {
	var dignitarieList ds.DignitarieList
	err := r.publishedWithLikes().Where("id > ?", id).Order("id").First(&dignitarieList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) && id > 0 {
		return r.GetNextDignitarieList(0)
	}
	return dignitarieList, wrapNotFound(err)
}

// GetDraftDignitarieList — черновик пользователя (не более одного) (ORM)
func (r *Repository) GetDraftDignitarieList(creatorID uint) (ds.DignitarieList, error) {
	var dignitarieList ds.DignitarieList
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&dignitarieList).Error
	return dignitarieList, wrapNotFound(err)
}

// CreateDraftDignitarieList — кнопка «Далее»: создание черновика с названием, url фото и видео (ORM)
func (r *Repository) CreateDraftDignitarieList(creatorID uint, name, imageURL, videoURL string) error {
	dignitarieList := ds.DignitarieList{
		Name:      name,
		ImageURL:  imageURL,
		VideoURL:  videoURL,
		Status:    ds.StatusDraft,
		CreatorID: creatorID,
	}
	return r.db.Create(&dignitarieList).Error
}

// PublishDignitarieList — кнопка «Опубликовать»: заполнение необязательных полей
// черновика и смена статуса на «опубликован» (ORM)
func (r *Repository) PublishDignitarieList(creatorID uint, description string, startDate, endDate time.Time) error {
	draft, err := r.GetDraftDignitarieList(creatorID)
	if err != nil {
		return err
	}

	now := time.Now()
	return r.db.Model(&draft).Updates(map[string]any{
		"description":  description,
		"start_date":   startDate,
		"end_date":     endDate,
		"status":       ds.StatusPublished,
		"published_at": &now,
	}).Error
}

// DeleteDignitarieList — логическое удаление услуги: SQL-запрос UPDATE без ORM,
// результат читается через курсор (sql.Rows)
func (r *Repository) DeleteDignitarieList(ctx context.Context, id uint) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	rows, err := sqlDB.QueryContext(ctx,
		"UPDATE dignitarie_lists SET status = $1 WHERE id = $2 AND status <> $1 RETURNING id",
		ds.StatusDeleted, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	// курсор: перебираем строки, изменённые запросом
	deleted := 0
	for rows.Next() {
		var deletedID uint
		if err := rows.Scan(&deletedID); err != nil {
			return err
		}
		deleted++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if deleted == 0 {
		return fmt.Errorf("услуга id=%d: %w", id, ErrNotFound)
	}
	return nil
}

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

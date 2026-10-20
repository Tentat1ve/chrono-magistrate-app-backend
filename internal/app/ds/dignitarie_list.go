package ds

import "time"

// Статусы услуги
const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

// DignitarieList — услуга: сановник (визирь, консул, воевода) с датами пребывания в должности.
// Обязательные поля заполняются по кнопке «Далее» (название, url фото и видео),
// необязательных полей четыре: краткое описание, две даты по теме и дата формирования.
// У пользователя не более одного черновика — частичный уникальный индекс idx_one_draft_per_creator.
type DignitarieList struct {
	ID          uint       `gorm:"primaryKey"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Description *string    `gorm:"type:text"`
	Status      string     `gorm:"type:varchar(20);not null;default:'черновик';uniqueIndex:idx_one_draft_per_creator,where:status = 'черновик'"`
	ImageURL    string     `gorm:"type:varchar(500);not null"`
	VideoURL    string     `gorm:"type:varchar(500);not null"`
	StartDate   *time.Time `gorm:"type:date"` // поле по теме 1: дата вступления в должность
	EndDate     *time.Time `gorm:"type:date"` // поле по теме 2: дата оставления должности
	CreatedAt   time.Time  `gorm:"not null"`  // дата создания
	PublishedAt *time.Time // дата формирования (публикации)
	CreatorID   uint       `gorm:"not null;uniqueIndex:idx_one_draft_per_creator"`
	Creator     User       `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`

	Likes []DignitarieListLike `gorm:"foreignKey:DignitarieListID"`
}

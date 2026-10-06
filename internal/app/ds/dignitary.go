package ds

import "time"

// Статусы сановника
const (
	DignitaryStatusDraft     = "черновик"
	DignitaryStatusPublished = "опубликован"
	DignitaryStatusDeleted   = "удален"
)

// Dignitary — услуга: сановник (визирь, консул, воевода) с годами пребывания в должности.
// Годы до н.э. хранятся отрицательными числами (-63 = 63 г. до н.э.).
// У пользователя не более одного черновика — частичный уникальный индекс idx_one_draft_per_creator.
type Dignitary struct {
	ID          uint       `gorm:"primaryKey"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Office      string     `gorm:"type:varchar(255)"`
	Description string     `gorm:"type:text"`
	Status      string     `gorm:"type:varchar(20);not null;default:'черновик';uniqueIndex:idx_one_draft_per_creator,where:status = 'черновик'"`
	ImageURL    string     `gorm:"type:varchar(500)"`
	VideoURL    string     `gorm:"type:varchar(500)"`
	OfficeStart *int       // поле по теме 1: год вступления в должность
	OfficeEnd   *int       // поле по теме 2: год оставления должности
	CreatedAt   time.Time  `gorm:"not null"`         // дата создания
	PublishedAt *time.Time                           // дата формирования (публикации)
	CreatorID   uint       `gorm:"not null;uniqueIndex:idx_one_draft_per_creator"`
	Creator     User       `gorm:"foreignKey:CreatorID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`

	Likes []DignitaryLike `gorm:"foreignKey:DignitaryID"`
}

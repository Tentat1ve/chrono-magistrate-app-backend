package ds

import "time"

// User — пользователь (создатель сановников)
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Login        string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	FullName     string    `gorm:"type:varchar(100);not null"`
	PasswordHash string    `gorm:"type:varchar(60);not null;default:''"`
	CreatedAt    time.Time `gorm:"not null"`
}

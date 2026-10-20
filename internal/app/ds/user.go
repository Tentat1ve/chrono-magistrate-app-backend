package ds

// User — пользователь (создатель услуг)
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Login string `gorm:"type:varchar(50);not null;uniqueIndex"`
}

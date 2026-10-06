package ds

// DignitaryLike — лайк, связь м-м пользователь–сановник
type DignitaryLike struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_like_user_dignitary"`
	DignitaryID uint      `gorm:"not null;uniqueIndex:idx_like_user_dignitary"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	Dignitary   Dignitary `gorm:"foreignKey:DignitaryID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

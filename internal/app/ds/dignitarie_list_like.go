package ds

// DignitarieListLike — лайк, связь м-м пользователь–услуга
type DignitarieListLike struct {
	ID               uint           `gorm:"primaryKey"`
	UserID           uint           `gorm:"not null;uniqueIndex:idx_like_user_dignitarie_list"`
	DignitarieListID uint           `gorm:"not null;uniqueIndex:idx_like_user_dignitarie_list"`
	User             User           `gorm:"foreignKey:UserID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
	DignitarieList   DignitarieList `gorm:"foreignKey:DignitarieListID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT"`
}

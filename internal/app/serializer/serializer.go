// Package serializer — структуры запросов и ответов API (сериализация в JSON).
// Системные поля (id, статус, создатель, даты) в запросах отсутствуют —
// клиент не может их изменить, они вычисляются на сервере.
package serializer

import (
	"time"

	"awesomeProject/internal/app/ds"
)

// ---------- Сановники ----------

// DignitaryJSON — сановник в ответах API
type DignitaryJSON struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Office      string     `json:"office"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	ImageURL    string     `json:"image_url"`
	VideoURL    string     `json:"video_url"`
	OfficeStart *int       `json:"office_start"`
	OfficeEnd   *int       `json:"office_end"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at"`
	CreatorID   uint       `json:"creator_id"`
	LikesCount  int        `json:"likes_count"`
}

// DignitaryListItemJSON — элемент списка: признак 0/1, что текущий пользователь — создатель
type DignitaryListItemJSON struct {
	DignitaryJSON
	IsCreator int `json:"is_creator"`
}

// DignitaryFeedJSON — лента: признак 0/1, что текущий пользователь поставил лайк
type DignitaryFeedJSON struct {
	DignitaryJSON
	IsLiked int `json:"is_liked"`
}

// PublishDignitaryRequest — тело PUT публикации черновика
type PublishDignitaryRequest struct {
	Office      string `json:"office" binding:"required,max=255"`
	Description string `json:"description" binding:"required"`
	OfficeStart *int   `json:"office_start" binding:"required"`
	OfficeEnd   *int   `json:"office_end" binding:"required"`
}

// LikeRequest — тело POST лайка: 1 ставит лайк, 0 отменяет
type LikeRequest struct {
	Liked *int `json:"liked" binding:"required,oneof=0 1"`
}

// LikeResponse — ответ на лайк
type LikeResponse struct {
	DignitaryID uint `json:"dignitary_id"`
	IsLiked     int  `json:"is_liked"`
	LikesCount  int  `json:"likes_count"`
}

func DignitaryToJSON(d ds.Dignitary) DignitaryJSON {
	return DignitaryJSON{
		ID:          d.ID,
		Name:        d.Name,
		Office:      d.Office,
		Description: d.Description,
		Status:      d.Status,
		ImageURL:    d.ImageURL,
		VideoURL:    d.VideoURL,
		OfficeStart: d.OfficeStart,
		OfficeEnd:   d.OfficeEnd,
		CreatedAt:   d.CreatedAt,
		PublishedAt: d.PublishedAt,
		CreatorID:   d.CreatorID,
		LikesCount:  len(d.Likes),
	}
}

func DignitaryToListItem(d ds.Dignitary, currentUserID uint) DignitaryListItemJSON {
	return DignitaryListItemJSON{
		DignitaryJSON: DignitaryToJSON(d),
		IsCreator:     boolToInt(d.CreatorID == currentUserID),
	}
}

func DignitaryToFeed(d ds.Dignitary, currentUserID uint) DignitaryFeedJSON {
	return DignitaryFeedJSON{
		DignitaryJSON: DignitaryToJSON(d),
		IsLiked:       boolToInt(IsLikedBy(d, currentUserID)),
	}
}

func IsLikedBy(d ds.Dignitary, userID uint) bool {
	for _, like := range d.Likes {
		if like.UserID == userID {
			return true
		}
	}
	return false
}

// ---------- Пользователи ----------

// UserJSON — пользователь в ответах API (без пароля)
type UserJSON struct {
	ID        uint      `json:"id"`
	Login     string    `json:"login"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest — тело POST регистрации
type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=50"`
	FullName string `json:"full_name" binding:"required,max=100"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginRequest — тело POST аутентификации
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func UserToJSON(u ds.User) UserJSON {
	return UserJSON{ID: u.ID, Login: u.Login, FullName: u.FullName, CreatedAt: u.CreatedAt}
}

// ErrorJSON — описание ошибки
type ErrorJSON struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

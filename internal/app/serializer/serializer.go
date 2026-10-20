// Package serializer — структуры запросов и ответов API (сериализация в JSON).
// В запросах нет системных полей (id, статус, создатель, даты создания и формирования) —
// клиент не может их изменить, они вычисляются на сервере.
// В ответах только те поля, которые нужны странице.
package serializer

import (
	"time"

	"awesomeProject/internal/app/ds"
)

const dateLayout = "2006-01-02"

// ---------- Услуги ----------

// DignitarieListItemJSON — элемент списка (плитка), 7 полей.
// is_creator: 1, если создатель услуги — текущий пользователь.
type DignitarieListItemJSON struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	ImageURL   string  `json:"image_url"`
	StartDate  *string `json:"start_date"`
	EndDate    *string `json:"end_date"`
	LikesCount int     `json:"likes_count"`
	IsCreator  int     `json:"is_creator"`
}

// DignitarieListFeedJSON — лента, 10 полей: видео и фото, название, краткое описание,
// два поля по теме, число лайков и два признака 0/1.
// is_liked: 1, если текущий пользователь поставил лайк; is_creator: 1, если он создатель.
type DignitarieListFeedJSON struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	LikesCount  int     `json:"likes_count"`
	IsLiked     int     `json:"is_liked"`
	IsCreator   int     `json:"is_creator"`
}

// DignitarieListJSON — черновик, результат добавления и публикации, 8 полей
type DignitarieListJSON struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Status      string  `json:"status"`
}

// PublishDignitarieListRequest — тело PUT публикации черновика; даты в формате ГГГГ-ММ-ДД
type PublishDignitarieListRequest struct {
	Description string `json:"description" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}

// LikeRequest — тело POST лайка: 1 ставит лайк, 0 отменяет
type LikeRequest struct {
	Liked *int `json:"liked" binding:"required,oneof=0 1"`
}

// LikeResponse — ответ на лайк
type LikeResponse struct {
	DignitarieListID uint `json:"dignitarie_list_id"`
	IsLiked          int  `json:"is_liked"`
	LikesCount       int  `json:"likes_count"`
}

func DignitarieListToItem(d ds.DignitarieList, currentUserID uint) DignitarieListItemJSON {
	return DignitarieListItemJSON{
		ID:         d.ID,
		Name:       d.Name,
		ImageURL:   d.ImageURL,
		StartDate:  formatDate(d.StartDate),
		EndDate:    formatDate(d.EndDate),
		LikesCount: len(d.Likes),
		IsCreator:  boolToInt(d.CreatorID == currentUserID),
	}
}

func DignitarieListToFeed(d ds.DignitarieList, currentUserID uint) DignitarieListFeedJSON {
	return DignitarieListFeedJSON{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		ImageURL:    d.ImageURL,
		VideoURL:    d.VideoURL,
		StartDate:   formatDate(d.StartDate),
		EndDate:     formatDate(d.EndDate),
		LikesCount:  len(d.Likes),
		IsLiked:     boolToInt(IsLikedBy(d, currentUserID)),
		IsCreator:   boolToInt(d.CreatorID == currentUserID),
	}
}

func DignitarieListToJSON(d ds.DignitarieList) DignitarieListJSON {
	return DignitarieListJSON{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		ImageURL:    d.ImageURL,
		VideoURL:    d.VideoURL,
		StartDate:   formatDate(d.StartDate),
		EndDate:     formatDate(d.EndDate),
		Status:      d.Status,
	}
}

func IsLikedBy(d ds.DignitarieList, userID uint) bool {
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
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

// RegisterRequest — тело POST регистрации
type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

// LoginRequest — тело POST аутентификации
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func UserToJSON(u ds.User) UserJSON {
	return UserJSON{ID: u.ID, Login: u.Login}
}

// ErrorJSON — описание ошибки
type ErrorJSON struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(dateLayout)
	return &s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

package handler

import (
	"net/http"
	"time"

	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// Фото и видео по умолчанию хранятся на SSR-сервере вместе с иконками
const (
	defaultImageURL = "/static/img/default_dignitarie_list.svg"
	defaultVideoURL = "/static/img/default_dignitarie_list.mp4"
)

const catalogURL = "/dignitarie_lists/catalog"

// currentUserID — текущий пользователь (создатель) до появления авторизации
const currentUserID uint = 1

type Handler struct {
	Repository *repository.Repository
	minioURL   string
	httpClient *http.Client
}

func NewHandler(r *repository.Repository, minioURL string) *Handler {
	return &Handler{
		Repository: r,
		minioURL:   minioURL,
		httpClient: &http.Client{Timeout: 700 * time.Millisecond},
	}
}

// RegisterHandler — 6 HTTP-методов: 3 GET, 2 POST через ORM и 1 POST удаления через SQL
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/dignitarie_lists/feed", h.GetDignitarieListFeed)
	router.GET("/dignitarie_lists/draft", h.GetDignitarieListDraft)
	router.GET("/dignitarie_lists/catalog", h.GetDignitarieListCatalog)
	router.POST("/dignitarie_lists/draft", h.CreateDignitarieListDraft)
	router.POST("/dignitarie_lists/draft/publish", h.PublishDignitarieListDraft)
	router.POST("/dignitarie_lists/delete", h.DeleteDignitarieList)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// несуществующие адреса перенаправляются на плитку вместо ответа 404
	router.NoRoute(func(ctx *gin.Context) {
		ctx.Redirect(http.StatusSeeOther, catalogURL)
	})
}

// mediaURL возвращает url, если файл по нему доступен, иначе url файла по умолчанию
func (h *Handler) mediaURL(url, fallback string) string {
	if url == "" {
		return fallback
	}
	resp, err := h.httpClient.Head(url)
	if err != nil {
		return fallback
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fallback
	}
	return url
}

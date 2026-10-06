package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// Фото и видео по умолчанию хранятся на SSR-сервере вместе с иконками
const (
	defaultImageURL = "/static/img/default_dignitary.svg"
	defaultVideoURL = "/static/img/default_dignitary.mp4"
)

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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/dignitaries", h.GetDignitaries)
	router.GET("/dignitary_feed", h.GetDignitaryFeed)
	router.GET("/dignitary_feed/:id", h.GetDignitaryFeed)
	router.GET("/dignitary_draft", h.GetDignitaryDraft)
	router.POST("/dignitary_draft", h.CreateDignitaryDraft)
	router.POST("/dignitary_draft/publish", h.PublishDignitaryDraft)
	router.POST("/dignitaries/:id/delete", h.DeleteDignitary)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.SetFuncMap(template.FuncMap{"formatYear": formatYear})
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")
}

// formatYear выводит год с учётом эры: -63 → «63 г. до н.э.», 1565 → «1565 г.»
func formatYear(year *int) string {
	if year == nil {
		return "—"
	}
	if *year < 0 {
		return fmt.Sprintf("%d г. до н.э.", -*year)
	}
	return fmt.Sprintf("%d г.", *year)
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

func (h *Handler) errorPage(ctx *gin.Context, status int, message string) {
	ctx.HTML(status, "error.html", gin.H{"status": status, "message": message})
}

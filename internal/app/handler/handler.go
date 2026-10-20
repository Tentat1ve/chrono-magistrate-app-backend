package handler

import (
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler — 3 GET-запроса: лента, черновик (добавление), плитка
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/dignitarie_lists/feed", h.GetDignitarieListFeed)
	router.GET("/dignitarie_lists/draft", h.GetDignitarieListDraft)
	router.GET("/dignitarie_lists/catalog", h.GetDignitarieListCatalog)
}

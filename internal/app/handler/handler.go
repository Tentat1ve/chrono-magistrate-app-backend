package handler

import (
	"net/http"

	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler — маршруты веб-сервиса, все начинаются с /api
func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	// Домен «сановники» (услуги)
	dignitaries := api.Group("/dignitaries")
	dignitaries.GET("", h.GetDignitaries)
	dignitaries.POST("", h.CreateDignitary)
	dignitaries.GET("/feed", h.GetDignitaryFeed)
	dignitaries.GET("/feed/:id", h.GetDignitaryFeed)
	dignitaries.GET("/draft", h.GetDignitaryDraft)
	dignitaries.PUT("/draft/publish", h.PublishDignitaryDraft)
	dignitaries.DELETE("/:id", h.DeleteDignitary)
	dignitaries.POST("/:id/like", h.LikeDignitary)

	// Домен «пользователи»
	users := api.Group("/users")
	users.POST("/register", h.RegisterUser)
	users.POST("/login", h.LoginUser)
	users.POST("/logout", h.LogoutUser)
}

func (h *Handler) errorJSON(ctx *gin.Context, status int, description string) {
	ctx.AbortWithStatusJSON(status, serializer.ErrorJSON{Status: "error", Description: description})
}

func (h *Handler) internalError(ctx *gin.Context, err error) {
	logrus.Error(err)
	h.errorJSON(ctx, http.StatusInternalServerError, "внутренняя ошибка сервера")
}

package handler

import (
	"net/http"

	"awesomeProject/internal/app/config"
	"awesomeProject/internal/app/redis"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWT        config.JWTConfig
}

func NewHandler(r *repository.Repository, redisClient *redis.Client, jwtConfig config.JWTConfig) *Handler {
	return &Handler{Repository: r, Redis: redisClient, JWT: jwtConfig}
}

// RegisterHandler — 10 методов веб-сервиса, все адреса начинаются с /api.
// Чтение (список, лента) и вход доступны гостю, методы создателя требуют токен.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	// Домен «услуги» — 7 методов
	dignitarieLists := api.Group("/dignitarie_lists")
	dignitarieLists.GET("", h.WithOptionalAuth, h.GetDignitarieLists)
	dignitarieLists.GET("/feed", h.WithOptionalAuth, h.GetDignitarieListFeed)
	dignitarieLists.GET("/draft", h.WithAuthCheck, h.GetDignitarieListDraft)
	dignitarieLists.POST("", h.WithAuthCheck, h.CreateDignitarieList)
	dignitarieLists.PUT("/draft/publish", h.WithAuthCheck, h.PublishDignitarieListDraft)
	dignitarieLists.DELETE("/:id", h.WithAuthCheck, h.DeleteDignitarieList)
	dignitarieLists.POST("/:id/like", h.WithAuthCheck, h.LikeDignitarieList)

	// Домен «пользователи» — 3 метода
	users := api.Group("/users")
	users.POST("/register", h.RegisterUser)
	users.POST("/login", h.LoginUser)
	users.POST("/logout", h.WithAuthCheck, h.LogoutUser)
}

func (h *Handler) errorJSON(ctx *gin.Context, status int, description string) {
	ctx.AbortWithStatusJSON(status, serializer.ErrorJSON{Status: "error", Description: description})
}

func (h *Handler) internalError(ctx *gin.Context, err error) {
	logrus.Error(err)
	h.errorJSON(ctx, http.StatusInternalServerError, "внутренняя ошибка сервера")
}

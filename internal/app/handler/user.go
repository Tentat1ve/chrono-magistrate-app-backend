package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser godoc
// @Summary      Регистрация
// @Description  Создаёт пользователя; пароль хранится в виде bcrypt-хеша и в ответе не возвращается. Доступна гостю.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        body  body      serializer.RegisterRequest  true  "Логин (3–50 символов) и пароль (от 6 символов)"
// @Success      201   {object}  serializer.UserJSON
// @Failure      400   {object}  serializer.ErrorJSON
// @Failure      409   {object}  serializer.ErrorJSON
// @Router       /users/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req serializer.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "укажите login (3–50 символов) и password (от 6 символов)")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	user := ds.User{Login: strings.TrimSpace(req.Login), PasswordHash: string(hash)}
	if err := h.Repository.CreateUser(&user); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorJSON(ctx, http.StatusConflict, "пользователь с таким логином уже существует")
			return
		}
		h.internalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, serializer.UserToJSON(user))
}

// LoginUser godoc
// @Summary      Аутентификация
// @Description  Проверяет логин и пароль и выдаёт JWT. Токен передаётся в заголовке Authorization: Bearer <токен>. Доступна гостю.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        body  body      serializer.LoginRequest  true  "Логин и пароль"
// @Success      200   {object}  serializer.LoginResponse
// @Failure      400   {object}  serializer.ErrorJSON
// @Failure      401   {object}  serializer.ErrorJSON
// @Router       /users/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var req serializer.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "укажите login и password")
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err == nil {
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	}
	if err != nil {
		h.errorJSON(ctx, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	token, _, err := h.newToken(user.ID)
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.JWT.ExpiresIn.Seconds()),
	})
}

// LogoutUser godoc
// @Summary      Деавторизация
// @Description  Отзывает токен: его идентификатор попадает в чёрный список Redis до конца срока жизни токена, после этого токен не принимается. Требует авторизации.
// @Tags         Пользователи
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  serializer.StatusResponse
// @Failure      401  {object}  serializer.ErrorJSON
// @Router       /users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	claims := currentClaims(ctx)

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl > 0 {
		if err := h.Redis.AddToBlacklist(ctx.Request.Context(), claims.ID, ttl); err != nil {
			h.internalError(ctx, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, serializer.StatusResponse{Status: "ok", Message: "токен отозван"})
}

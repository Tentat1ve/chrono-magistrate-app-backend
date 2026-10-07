package handler

import (
	"errors"
	"net/http"
	"strings"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser — POST /api/users/register
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req serializer.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "укажите login (3–50 символов), full_name и password (от 6 символов)")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	user := ds.User{
		Login:        strings.TrimSpace(req.Login),
		FullName:     strings.TrimSpace(req.FullName),
		PasswordHash: string(hash),
	}
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

// LoginUser — POST /api/users/login
// Заглушка до ЛР4: проверяет логин и пароль, но сессию/токен не выдаёт.
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

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "заглушка: токен будет выдаваться в ЛР4",
		"user":    serializer.UserToJSON(user),
	})
}

// LogoutUser — POST /api/users/logout — заглушка до ЛР4
func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "заглушка: деавторизация будет в ЛР4"})
}

package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"awesomeProject/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	jwtPrefix = "Bearer "

	ctxClaims = "jwt_claims" // содержимое токена в контексте запроса
)

var errUnauthorized = errors.New("требуется авторизация")

// newToken создаёт JWT для пользователя: подпись HS256, срок жизни из конфигурации
func (h *Handler) newToken(userID uint) (string, time.Time, error) {
	expiresAt := time.Now().Add(h.JWT.ExpiresIn)
	claims := ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(), // идентификатор токена — по нему токен попадает в чёрный список
			Issuer:    "dignitarie_lists",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.JWT.Secret))
	return signed, expiresAt, err
}

// parseToken достаёт токен из заголовка Authorization, проверяет подпись, срок и чёрный список
func (h *Handler) parseToken(ctx *gin.Context) (*ds.JWTClaims, error) {
	header := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(header, jwtPrefix) {
		return nil, errUnauthorized
	}
	tokenString := strings.TrimPrefix(header, jwtPrefix)

	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(h.JWT.Secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, errUnauthorized
	}

	blacklisted, err := h.Redis.InBlacklist(ctx.Request.Context(), claims.ID)
	if err != nil {
		return nil, err
	}
	if blacklisted { // токен отозван через logout
		return nil, errUnauthorized
	}
	return claims, nil
}

// WithAuthCheck — middleware методов создателя: без действующего токена ответ 401
func (h *Handler) WithAuthCheck(ctx *gin.Context) {
	claims, err := h.parseToken(ctx)
	if errors.Is(err, errUnauthorized) {
		h.errorJSON(ctx, http.StatusUnauthorized, "требуется авторизация: передайте токен в заголовке Authorization (Bearer)")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.Set(ctxClaims, claims)
	ctx.Next()
}

// WithOptionalAuth — middleware чтения: гость проходит без токена,
// авторизованному пользователю в ответ добавляются признаки is_creator и is_liked
func (h *Handler) WithOptionalAuth(ctx *gin.Context) {
	claims, err := h.parseToken(ctx)
	if err == nil {
		ctx.Set(ctxClaims, claims)
	} else if !errors.Is(err, errUnauthorized) {
		logrus.Error(err)
	}
	ctx.Next()
}

// currentUserID — ID авторизованного пользователя из токена, 0 для гостя
func currentUserID(ctx *gin.Context) uint {
	if claims, ok := ctx.Get(ctxClaims); ok {
		return claims.(*ds.JWTClaims).UserID
	}
	return 0
}

// currentClaims — токен текущего запроса (после WithAuthCheck)
func currentClaims(ctx *gin.Context) *ds.JWTClaims {
	claims, _ := ctx.Get(ctxClaims)
	c, _ := claims.(*ds.JWTClaims)
	return c
}

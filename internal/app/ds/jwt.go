package ds

import "github.com/golang-jwt/jwt/v5"

// JWTClaims — содержимое токена авторизации: стандартные поля (срок, выпуск, идентификатор)
// и ID пользователя, которому он выдан
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

package api

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/limbo/discipline/pkg/entity"
)

type JWTServiceI interface {
	GenerateToken(user *entity.User) (string, error)
	ParseToken(tokenString string) (*JWTClaims, error)
}

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

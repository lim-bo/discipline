package jwtservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/limbo/discipline/internal/api"
	errorvalues "github.com/limbo/discipline/internal/error_values"
	"github.com/limbo/discipline/pkg/entity"
)

var (
	tokenTTL = time.Hour
)

type JWTService struct {
	secret []byte
}

func New(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
	}
}

func (s *JWTService) GenerateToken(user *entity.User) (string, error) {
	expTime := time.Now().Add(tokenTTL)
	claims := &api.JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseToken(tokenString string) (*api.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &api.JWTClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, errors.New("token parsing error: " + err.Error())
	}
	claims, ok := token.Claims.(*api.JWTClaims)
	if !ok || !token.Valid {
		return nil, errorvalues.ErrInvalidToken
	}
	return claims, nil
}

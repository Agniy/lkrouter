package service

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"lkrouter/config"
)

type JwtService struct {
	secret string
}

func NewJwtService() *JwtService {
	cfg := config.GetConfig()
	return &JwtService{
		secret: cfg.JwtSecret,
	}
}

func (j *JwtService) GenerateToken() string {
	tokenBuilder := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": "123",
	})

	tokenString, err := tokenBuilder.SignedString([]byte(j.secret))
	if err != nil {
		fmt.Errorf("Error generating token: %v", err)
	}
	return tokenString
}

func (j *JwtService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %v", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("Error parsing token")
}

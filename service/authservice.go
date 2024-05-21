package service

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"lkrouter/pkg/mongodb/mrequests"
	"strings"
)

type AuthService struct {
	jwtService *JwtService
}

func NewAuthService() *AuthService {
	jwtService := NewJwtService()
	return &AuthService{
		jwtService: jwtService,
	}
}

func (a *AuthService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	return a.jwtService.ParseToken(tokenString)
}

func (a *AuthService) CheckRoomPermission(c *gin.Context, room string) bool {
	authHeader := c.Request.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimLeft(authHeader, "Bearer"))

	claims, err := a.ParseToken(token)
	if err != nil {
		return false
	}
	uid, ok := claims["uid"]
	if !ok {
		return false
	}

	// get room from db and check if uid has permission
	call, err := mrequests.GetCallByRoom(room)
	if err != nil {
		return false
	}

	// check if uid is initiator
	initiator := ""
	if call["initiator"] != nil {
		initiator = call["initiator"].(string)
	}

	if initiator == uid {
		return true
	}

	return false
}

package helpers

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"os"
	"project/cmd/database/model"
	"strconv"
	"time"
)

type TokenRequest struct {
	Token string `json:"token"`
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId uint `json:"id"`
}

func GenerateSessionToken(user *model.User) (string, error) {
	ttl, err := strconv.Atoi(os.Getenv("TTL_HOUR"))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})

	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

func ParseUserToken(accessToken string) (uint, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenCliams")
	}

	return claims.UserId, nil
}

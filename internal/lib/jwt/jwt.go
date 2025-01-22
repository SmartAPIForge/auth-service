package jwt

import (
	"auth-service/internal/domain/models"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type TokenPayload struct {
	Type  string
	Uid   int64
	Email string
	Role  int64
	Exp   int64
}

func NewToken(
	user models.User,
	duration time.Duration,
	tokenType string,
) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["type"] = tokenType
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte("do_not_forget_to_hide_please"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*TokenPayload, error) {
	var payload TokenPayload

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte("do_not_forget_to_hide_please"), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		payload = TokenPayload{
			Type:  claims["type"].(string),
			Uid:   int64(claims["uid"].(float64)),
			Email: claims["email"].(string),
			Role:  int64(claims["role"].(float64)),
			Exp:   int64(claims["exp"].(float64)),
		}
		return &payload, nil
	}

	return nil, errors.New("invalid token")
}

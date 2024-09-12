package internal

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateJwtToken(username string, role string, secret string) (string, error) {
	var secretKey = []byte(secret)
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"iss": "go-api-gateway",
		"aud": role,
		"exp": time.Now().Add(time.Hour).Unix(), // 1 Hour of expiration time
		"iat": time.Now().Unix(),
	})
	tokenStr, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func ValidJwtToken(tokenStr string, secret string) (jwt.Claims, error) {
	var secretKey = []byte(secret)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, err
	}
	isExpired := time.Now().UnixMilli() > exp.UnixMilli()
	if isExpired {
		return nil, errors.New("token is expired")
	}
	return token.Claims, nil
}

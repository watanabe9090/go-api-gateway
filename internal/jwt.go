package internal

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("your-secret-key")

func CreateJwtToken(username string, role string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"iss": "fortress",
		"aud": role,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func ValidJwtToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {

	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, err
	}
	isExpired := time.Now().UnixMilli() > exp.UnixMilli()
	if isExpired {
		return nil, err
	}
	return token.Claims, nil
}

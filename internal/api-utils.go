package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type APIResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func HttpReply(w http.ResponseWriter, statusCode int, body *APIResponse) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}

func GetBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	//ToDo: implement startsWith instead of Contains
	if authHeader == "" || !strings.Contains(strings.ToLower(authHeader), "bearer ") {
		return "", errors.New("")
	}
	tokenStr := authHeader[7:]
	return tokenStr, nil
}

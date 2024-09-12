package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/watanabe9090/cerberus/cmd/accounts"
	"github.com/watanabe9090/cerberus/cmd/tokens"
	"github.com/watanabe9090/cerberus/internal"
)

type AuthHandler struct {
	accRepo   *accounts.AccountsRepository
	tokenRepo *tokens.TokensRepository
	props     *internal.Properties
}

type DTONewToken struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(db *sql.DB, props *internal.Properties) *AuthHandler {
	accRepo := accounts.NewAccountsRepository(db)
	tokenRepo := tokens.NewTokensRepository(db)
	accRepo.InitUsersTable()
	tokenRepo.InitTokenTable()

	return &AuthHandler{
		accRepo:   accRepo,
		tokenRepo: tokenRepo,
		props:     props,
	}
}

func (h *AuthHandler) HandleNewToken(w http.ResponseWriter, r *http.Request) {
	var body DTONewToken
	json.NewDecoder(r.Body).Decode(&body)

	/* Check if the user exists by their username */
	acc, _ := h.accRepo.GetByUsername(body.Username)
	if acc == nil {
		internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
			Message: "user not found",
			Data:    nil,
		})
		return
	}

	/* Check if the password matches
	* I know it's not ideal to show the excat error
	* for security reasons during authentication,
	* but for now, I'll do it to be precise.
	* We can easily change this later.
	*
	* ToDo: Use bcrypt to hash the password, REF: https://gowebexamples.com/password-hashing/
	 */
	if acc.Password != body.Password {
		log.Printf("wrong password for %s \n", body.Username)
		internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
			Message: "wrong password",
			Data:    nil,
		})
		return
	}

	/* ToDo: Implement the refresh token (not needed at the moment)
	*  This will allow users to obtain a new token without re-authenticating
	 */
	token, err := internal.CreateJwtToken(acc.Username, acc.Role, h.props.JWT.Secret)
	if err != nil {
		log.Println(err.Error())
		internal.HttpReply(w, http.StatusInternalServerError, &internal.APIResponse{
			Message: "could not generate JWT token",
			Data:    err.Error(),
		})
		return
	}
	err = h.tokenRepo.SaveToken(acc.Username, token)
	if err != nil {
		log.Println(err.Error())
		internal.HttpReply(w, http.StatusInternalServerError, &internal.APIResponse{
			Message: "could not save the JWT token",
			Data:    err.Error(),
		})
		return
	}
	w.Header().Add("X-Auth-Username", acc.Username)
	w.Header().Add("X-Auth-Role", acc.Role)
	internal.HttpReply(w, http.StatusOK, &internal.APIResponse{
		Message: "OK",
		Data:    token,
	})
}

func (h *AuthHandler) HandleInvalidateToken(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := internal.GetBearerToken(r)
	if err != nil {
		log.Println("no bearer token in Authorization header")
		internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
			Message: "no bearer token in Authorization header",
			Data:    err.Error(),
		})
		return
	}
	savedTokens, err := h.tokenRepo.GetByToken(tokenStr)
	if len(savedTokens) == 0 {
		internal.HttpReply(w, http.StatusBadRequest, &internal.APIResponse{
			Message: "token not found",
			Data:    nil,
		})
		return
	}
	claims, err := internal.ValidJwtToken(tokenStr, h.props.JWT.Secret)
	if err != nil {
		internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
			Message: "invalid JWT token",
			Data:    err.Error(),
		})
		return
	}
	sub, err := claims.GetSubject()
	if err != nil {
		internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
			Message: "invalid JWT token",
			Data:    err.Error(),
		})
		return
	}
	err = h.tokenRepo.UpdateTokensState(sub, tokenStr, "INVALID")
	if err != nil {
		internal.HttpReply(w, http.StatusInternalServerError, &internal.APIResponse{
			Message: "could not invalidate JWT token",
			Data:    err.Error(),
		})
		return
	}
	internal.HttpReply(w, http.StatusNoContent, &internal.APIResponse{
		Message: "OK",
		Data:    nil,
	})
}

func (h *AuthHandler) HandleForward(w http.ResponseWriter, r *http.Request) {
	requestedPrefix := strings.Replace(r.URL.Path, "/api/v1", "", 1)
	log.Println(requestedPrefix)

	// Find the right context in properties.yml
	var route *internal.APIRoute
	for _, value := range h.props.APIs {
		if strings.Contains(requestedPrefix, value.Prefix) {
			route = &value
			break
		}
	}
	if route == nil {
		log.Printf("Cloud not found the route %s\n", requestedPrefix)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Find if the route exits for the right METHOD
	var permission *internal.APIRoutePermission
	for _, value := range route.Routes {
		// ToDo: Apply startsWith insted of contains
		if value.Method == r.Method &&
			strings.Contains(requestedPrefix, value.Route) {
			permission = &value
			break
		}
	}
	if permission == nil {
		log.Printf("Cloud not found the route for %s and method %s\n", requestedPrefix, r.Method)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check the bearer tokenStr
	tokenStr, err := internal.GetBearerToken(r)
	if err != nil {
		log.Println("Cloud not found the Bearer token in Headers")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check fi the bearer token is active
	tokens, err := h.tokenRepo.GetByToken(tokenStr)
	if err != nil {
		log.Printf("Cloud not found the Bearer token %s", tokenStr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if len(tokens) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
	}
	// token := tokenStr[0]
	fmt.Println(tokenStr[0])

	claims, err := internal.ValidJwtToken(tokenStr, h.props.JWT.Secret)
	if err != nil {
		log.Printf("")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println(claims)

	fmt.Println(route)
	fmt.Println(permission)
	url := route.Host + permission.Route

	fmt.Println(url)

	proxyReq, err := http.NewRequest(permission.Method, url, r.Body)
	if err != nil {
		log.Println(err.Error())
	}
	sub, err := claims.GetSubject()
	if err != nil {
		log.Println(err.Error())
	}
	// aud, err := claims.GetAudience()
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	proxyReq.Header.Add("X-Auth-Username", sub)
	proxyReq.Header.Add("X-Auth-Role", "ADMIN")
	proxyReq.Header.Add("X-Auth-Token", tokenStr)
	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}
	fmt.Println(proxyReq.URL)
	client := &http.Client{}
	proxyRes, err := client.Do(proxyReq)
	defer proxyRes.Body.Close()

	fmt.Print(proxyRes)
	for header, values := range proxyRes.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.WriteHeader(proxyRes.StatusCode)
	io.Copy(w, proxyRes.Body)

}

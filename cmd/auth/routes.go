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
	requestedContext := strings.Replace(r.URL.Path, "/api/v1", "", 1)
	log.Printf("request being forward to: %s\n", requestedContext)

	// Find the right context in properties.yml
	var route *internal.APIRoute
	for _, value := range h.props.APIs {
		if strings.HasPrefix(requestedContext, value.Context) {
			route = &value
			break
		}
	}
	if route == nil {
		log.Printf("could not found the route %s\n", requestedContext)
		internal.HttpReply(w, http.StatusNotFound, &internal.APIResponse{
			Message: fmt.Sprintf("could not found the route %s\n", requestedContext),
			Data:    nil,
		})
		return
	}

	// Find if the route exits for the right METHOD
	var subRoute *internal.APIRoutePermission
	for _, value := range route.Routes {
		if value.Method == r.Method &&
			strings.HasPrefix(requestedContext, value.Route) {
			subRoute = &value
			break
		}
	}
	if subRoute == nil {
		log.Printf("could not found the route for %s and method %s\n", requestedContext, r.Method)
		internal.HttpReply(w, http.StatusNotFound, &internal.APIResponse{
			Message: fmt.Sprintf("could not found the route for %s and method %s\n", requestedContext, r.Method),
			Data:    nil,
		})
		return
	}

	// Creates the forward request
	fwdUrl := route.Host + subRoute.Route + "?" + r.URL.RawQuery
	log.Printf("Forward url: %s\n", fwdUrl)
	fwd, err := http.NewRequest(subRoute.Method, fwdUrl, r.Body)
	if err != nil {
		log.Println(err.Error())
		internal.HttpReply(w, http.StatusInternalServerError, &internal.APIResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	if subRoute.Role != "NONE" {
		// Check the bearer tokenStr
		tokenStr, err := internal.GetBearerToken(r)
		if err != nil {
			log.Println("could not found the Bearer token in Headers")
			internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
				Message: "could not found the Bearer token in Headers",
				Data:    nil,
			})
			return
		}
		// Check fi the bearer token is active
		tokens, err := h.tokenRepo.GetByToken(tokenStr)
		if err != nil || len(tokens) == 0 {
			log.Printf("could not found the Bearer token %s", tokenStr)
			internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
				Message: fmt.Sprintf("could not found the Bearer token %s", tokenStr),
				Data:    nil,
			})
			return
		}

		claims, err := internal.ValidJwtToken(tokenStr, h.props.JWT.Secret)
		if err != nil {
			log.Printf(err.Error())
			internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
		sub, err := claims.GetSubject()
		if err != nil {
			log.Println(err.Error())
			internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
		aud, err := claims.GetAudience()
		if err != nil {
			log.Println(err.Error())
			internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
		if len(aud) != 0 {
			fwd.Header.Add("X-Auth-Role", aud[0])
		} else {
			log.Printf("no aud for %s\n", sub)
			internal.HttpReply(w, http.StatusUnauthorized, &internal.APIResponse{
				Message: fmt.Sprintf("no aud for %s\n", sub),
				Data:    nil,
			})
			return
		}
		fwd.Header.Add("X-Auth-Username", sub)
	}

	for header, values := range r.Header {
		for _, value := range values {
			fwd.Header.Add(header, value)
		}
	}

	client := &http.Client{}
	fwdRes, err := client.Do(fwd)

	// ToDo: how to deal with connection refused - error 502 or 503
	if err != nil {
		log.Println(err.Error())
		internal.HttpReply(w, 502, &internal.APIResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
	defer fwdRes.Body.Close()

	for header, values := range fwdRes.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(fwdRes.StatusCode)
	io.Copy(w, fwdRes.Body)
}

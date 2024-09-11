package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/watanabe9090/cerberus/cmd/auth"
	"github.com/watanabe9090/cerberus/internal"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalln("No yaml file provide")
	}
	props := internal.ParseYml(args[0])
	mux := http.NewServeMux()
	db, err := internal.OpenPostgreSQLConnection(&props.DB)
	if err != nil {
		log.Fatalln("Could not open database connection")
	}
	authHand := auth.NewAuthHandler(db, &props)
	mux.HandleFunc("POST /api/v1/auth/token", authHand.HandleNewToken)
	mux.HandleFunc("POST /api/v1/auth/invalidate", authHand.HandleInvalidateToken)
	mux.HandleFunc("/api/v1/", authHand.HandleForward)
	fmt.Printf("Server running on port %d\n", props.Server.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", props.Server.Port), mux)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

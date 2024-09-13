package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	metrics := internal.InitPrometheusMetrics()
	authHand := auth.NewAuthHandler(db, &props)
	mux.HandleFunc("POST /api/v1/auth/token", metrics.RequestTimeMetric(authHand.HandleNewToken))
	mux.HandleFunc("POST /api/v1/auth/invalidate", metrics.RequestTimeMetric(authHand.HandleInvalidateToken))
	mux.HandleFunc("/api/v1/", metrics.RequestTimeMetric(authHand.HandleForward))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", metrics.RequestTimeMetric(func(w http.ResponseWriter, r *http.Request) {
		internal.HttpReply(w, http.StatusNotFound, &internal.APIResponse{
			Message: fmt.Sprintf("could not found the route %s", r.URL.String()),
			Data:    nil,
		})
	}))
	fmt.Printf("Server running on port %d\n", props.Server.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", props.Server.Port), mux)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

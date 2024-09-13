package internal

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Details in https://robert-scherbarth.medium.com/measure-request-duration-with-prometheus-and-golang-adc6f4ca05fe
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

type PrometheusMetrics struct {
	responseTimeHistogram *prometheus.HistogramVec
}

func InitPrometheusMetrics() *PrometheusMetrics {
	buckets := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	responseTimeHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Name:      "http_server_request_duration_seconds",
		Help:      "Histogram of response time for handler in seconds",
		Buckets:   buckets,
	}, []string{"route", "method", "status_code"})
	prometheus.MustRegister(responseTimeHistogram)
	return &PrometheusMetrics{
		responseTimeHistogram: responseTimeHistogram,
	}
}

func (m *PrometheusMetrics) RequestTimeMetric(f http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}
		f.ServeHTTP(&rec, r)
		duration := time.Since(start)
		fmt.Println(duration)
		m.responseTimeHistogram.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(rec.statusCode)).Observe(duration.Seconds())
	}
}

package metrics

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// AircraftSpottedTotal counts the total number of aircraft spotted, labelled by tail number.
	AircraftSpottedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geebee_aircraft_spotted_total",
			Help: "Total number of aircraft spotted, labelled by tail number.",
		},
		[]string{"tail_number"},
	)

	// APICallsTotal counts the total number of calls made to the upstream ADS-B API.
	APICallsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "geebee_api_calls_total",
		Help: "Total number of calls made to the upstream ADS-B API.",
	})

	// APIErrorsTotal counts the total number of errors returned by the upstream ADS-B API.
	APIErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "geebee_api_errors_total",
		Help: "Total number of errors returned by the upstream ADS-B API.",
	})

	// NotificationsSentTotal counts notifications sent, labelled by channel type.
	NotificationsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geebee_notifications_sent_total",
			Help: "Total number of notifications sent, labelled by channel (discord, slack, custom_url, terminal).",
		},
		[]string{"channel"},
	)
)

// StartServer starts the Prometheus metrics HTTP server on the given port.
func StartServer(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Prometheus metrics available at http://0.0.0.0%s/metrics", addr)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Failed to start metrics server on port %d: %v", port, err)
		}
	}()
}

package worker

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts an HTTP server to expose Prometheus metrics
func (w *Worker) StartMetricsServer(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := ":" + port
	log.Printf("[METRICS] Worker %s metrics server starting on %s", w.ID, addr)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("[ERROR] Metrics server failed: %v", err)
		}
	}()

	return nil
}

// ServeMetrics is a simple health check endpoint
func (w *Worker) ServeMetrics() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "Worker %s metrics\n", w.ID)
		fmt.Fprintf(rw, "Active jobs: %d/%d\n", w.GetJobCount(), w.Capacity)
	})
}

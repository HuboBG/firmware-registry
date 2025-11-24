package api

import (
	"firmware-registry-api/internal/api/handlers"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter wires HTTP routes to handlers.
func NewRouter(fh *handlers.FirmwareHandler, wh *handlers.WebhookHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handlers.Health)
	mux.Handle("/api/firmware/", fh)
	mux.Handle("/api/webhooks", wh)
	mux.Handle("/api/webhooks/", wh)

	// Swagger UI at /swagger/index.html
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}

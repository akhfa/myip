package main

import (
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"myip/docs"
	"myip/internal/config"
	"myip/internal/handlers"
)

// @title MyIP API
// @version 1.0
// @description A lightweight, high-performance HTTP service for detecting client IP addresses with comprehensive proxy header support.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

func setupRoutes() {
	http.HandleFunc("/", handlers.IPv4Handler)
	http.HandleFunc("/ipv6", handlers.IPv6Handler)
	http.HandleFunc("/info", handlers.InfoHandler)
	http.HandleFunc("/json", handlers.JSONHandler)
	http.HandleFunc("/headers", handlers.HeadersHandler)
	http.HandleFunc("/health", handlers.HealthHandler)
	http.Handle("/swagger/", httpSwagger.WrapHandler)
}

func createServer(cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:              cfg.GetAddr(),
		Handler:           nil, // Use default ServeMux
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func main() {

	cfg := config.Load()

	// Update Swagger host dynamically
	docs.SwaggerInfo.Host = cfg.Host

	setupRoutes()

	server := createServer(cfg)

	log.Printf("Server starting on port %s", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

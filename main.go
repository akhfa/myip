package main

import (
	"log"
	"net/http"
	"time"
	
	"myip/internal/config"
	"myip/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Set up routes
	http.HandleFunc("/", handlers.IPv4Handler)
	http.HandleFunc("/ipv6", handlers.IPv6Handler)
	http.HandleFunc("/info", handlers.InfoHandler)
	http.HandleFunc("/json", handlers.JSONHandler)
	http.HandleFunc("/headers", handlers.HeadersHandler)
	http.HandleFunc("/health", handlers.HealthHandler)
	
	server := &http.Server{
		Addr:           cfg.GetAddr(),
		Handler:        nil,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server starting on port %s", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

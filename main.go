package main

import (
	"log"
	"net/http"
	
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
	
	log.Printf("Server starting on port %s", cfg.Port)
	
	if err := http.ListenAndServe(cfg.GetAddr(), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

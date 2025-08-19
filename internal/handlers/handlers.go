package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	
	"myip/internal/ip"
	"myip/internal/models"
)

// IPv4Handler handles requests for IPv4 addresses only
func IPv4Handler(w http.ResponseWriter, r *http.Request) {
	ipv4 := ip.FindIPv4(r)
	
	if ipv4 == "" {
		http.Error(w, "No IPv4 address found", http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv4)
}

// IPv6Handler handles requests for IPv6 addresses only
func IPv6Handler(w http.ResponseWriter, r *http.Request) {
	ipv6 := ip.FindIPv6(r)
	
	if ipv6 == "" {
		http.Error(w, "No IPv6 address found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv6)
}

// InfoHandler provides detailed IP information in plain text
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	info := ip.GetInfo(r)
	
	w.Header().Set("Content-Type", "text/plain")
	
	fmt.Fprintf(w, "Your IP Address: %s\n", info.ClientIP)
	fmt.Fprintf(w, "Detection Method: %s\n", info.DetectedVia)
	fmt.Fprintf(w, "Is Private IP: %t\n", info.IsPrivateIP)
	fmt.Fprintf(w, "Behind Cloudflare: %t\n", info.IsCloudflare)
	
	if info.IPv4Address != "" {
		fmt.Fprintf(w, "IPv4 Address: %s\n", info.IPv4Address)
	}
	if info.IPv6Address != "" {
		fmt.Fprintf(w, "IPv6 Address: %s\n", info.IPv6Address)
	}
	
	fmt.Fprintf(w, "Timestamp: %s\n", info.Timestamp)
}

// JSONHandler provides comprehensive JSON response
func JSONHandler(w http.ResponseWriter, r *http.Request) {
	info := ip.GetInfo(r)
	
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

// HeadersHandler shows all HTTP headers and IP details for debugging
func HeadersHandler(w http.ResponseWriter, r *http.Request) {
	info := ip.GetInfo(r)
	
	w.Header().Set("Content-Type", "text/plain")
	
	fmt.Fprintf(w, "=== IP INFORMATION ===\n")
	fmt.Fprintf(w, "Client IP: %s\n", info.ClientIP)
	fmt.Fprintf(w, "Detection Method: %s\n", info.DetectedVia)
	fmt.Fprintf(w, "IPv4 Address: %s\n", info.IPv4Address)
	fmt.Fprintf(w, "IPv6 Address: %s\n", info.IPv6Address)
	fmt.Fprintf(w, "Is Private IP: %t\n", info.IsPrivateIP)
	fmt.Fprintf(w, "Behind Cloudflare: %t\n", info.IsCloudflare)
	fmt.Fprintf(w, "Timestamp: %s\n", info.Timestamp)
	
	fmt.Fprintf(w, "\n=== HTTP HEADERS ===\n")
	
	// Sort headers for consistent output
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", name, value)
		}
	}
	
	fmt.Fprintf(w, "\n=== CONNECTION INFO ===\n")
	fmt.Fprintf(w, "Remote Address: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "URL: %s\n", r.URL.String())
	fmt.Fprintf(w, "Protocol: %s\n", r.Proto)
}

// HealthHandler provides health check endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.NewHealthResponse("healthy")
	
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		return
	}
}
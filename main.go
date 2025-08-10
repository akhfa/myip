package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func isIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

func isIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

func ipv4Handler(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	
	if isIPv4(clientIP) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, clientIP)
	} else {
		http.Error(w, "Request is not from IPv4", http.StatusBadRequest)
	}
}

func ipv6Handler(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	
	if isIPv6(clientIP) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, clientIP)
	} else {
		http.Error(w, "Request is not from IPv6", http.StatusBadRequest)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func main() {
	http.HandleFunc("/", ipv4Handler)
	http.HandleFunc("/ipv6", ipv6Handler)
	http.HandleFunc("/health", healthHandler)
	
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
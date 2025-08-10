package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// IPInfo represents detailed information about the client's IP
type IPInfo struct {
	ClientIP      string `json:"client_ip"`
	DetectedVia   string `json:"detected_via"`
	IPv4Address   string `json:"ipv4_address"`
	IPv6Address   string `json:"ipv6_address"`
	IsPrivateIP   bool   `json:"is_private_ip"`
	IsCloudflare  bool   `json:"is_cloudflare"`
	UserAgent     string `json:"user_agent"`
	Timestamp     string `json:"timestamp"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Header priority order for IP detection
var headerPriority = []string{
	"CF-Connecting-IP",     // Cloudflare
	"True-Client-IP",       // Cloudflare Enterprise
	"X-Real-IP",           // nginx proxy/FastCGI
	"X-Forwarded-For",     // Standard proxy header
	"X-Client-IP",         // Apache mod_proxy_http
	"X-Cluster-Client-IP", // Cluster environments
	"X-Forwarded",         // Less common
	"Forwarded-For",       // Less common
	"Forwarded",           // Less common
}

// Private IP ranges (IPv4)
var privateIPRanges = []*net.IPNet{
	// RFC 1918
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	// RFC 3927
	parseCIDR("169.254.0.0/16"),
	// RFC 5735
	parseCIDR("127.0.0.0/8"),
}

// Private IPv6 ranges
var privateIPv6Ranges = []*net.IPNet{
	// RFC 4193 - Unique Local Addresses
	parseCIDR("fc00::/7"),
	// RFC 4291 - Link-Local
	parseCIDR("fe80::/10"),
	// RFC 4291 - Loopback
	parseCIDR("::1/128"),
}

func parseCIDR(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("Failed to parse CIDR %s: %v", cidr, err)
	}
	return network
}

// isValidIP checks if the given string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// isPrivateIP checks if the given IP address is in a private range
func isPrivateIP(ip string) bool {
	if ip == "" {
		return false
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check IPv4 private ranges
	if parsedIP.To4() != nil {
		for _, privateRange := range privateIPRanges {
			if privateRange.Contains(parsedIP) {
				return true
			}
		}
		return false
	}

	// Check IPv6 private ranges
	for _, privateRange := range privateIPv6Ranges {
		if privateRange.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// isCloudflareRequest checks if the request comes through Cloudflare
func isCloudflareRequest(r *http.Request) bool {
	return r.Header.Get("CF-Connecting-IP") != "" || 
		   r.Header.Get("CF-Ray") != "" ||
		   r.Header.Get("True-Client-IP") != ""
}

// extractClientIP extracts the client IP from request headers with detection method
func extractClientIP(r *http.Request) (string, string) {
	// Check headers in priority order
	for _, header := range headerPriority {
		value := r.Header.Get(header)
		if value != "" {
			// Handle comma-separated IPs (take the first valid one)
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if isValidIP(ip) {
					return ip, header
				}
			}
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, "RemoteAddr"
	}
	return host, "RemoteAddr"
}

// findIPv4 finds the first valid IPv4 address from the request
func findIPv4(r *http.Request) string {
	// Check headers in priority order
	for _, header := range headerPriority {
		value := r.Header.Get(header)
		if value != "" {
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if isValidIP(ip) {
					parsedIP := net.ParseIP(ip)
					if parsedIP != nil && parsedIP.To4() != nil {
						return ip
					}
				}
			}
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	
	if isValidIP(host) {
		parsedIP := net.ParseIP(host)
		if parsedIP != nil && parsedIP.To4() != nil {
			return host
		}
	}

	return ""
}

// findIPv6 finds the first valid IPv6 address from the request
func findIPv6(r *http.Request) string {
	// Check headers in priority order
	for _, header := range headerPriority {
		value := r.Header.Get(header)
		if value != "" {
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if isValidIP(ip) {
					parsedIP := net.ParseIP(ip)
					if parsedIP != nil && parsedIP.To4() == nil {
						return ip
					}
				}
			}
		}
	}

	// Fall back to RemoteAddr (handle bracketed IPv6)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	
	if isValidIP(host) {
		parsedIP := net.ParseIP(host)
		if parsedIP != nil && parsedIP.To4() == nil {
			return host
		}
	}

	return ""
}

// removeDuplicates removes duplicate strings from a slice while preserving order
func removeDuplicates(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}
	
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// getIPInfo gets comprehensive IP information
func getIPInfo(r *http.Request) *IPInfo {
	clientIP, detectedVia := extractClientIP(r)
	ipv4 := findIPv4(r)
	ipv6 := findIPv6(r)
	
	return &IPInfo{
		ClientIP:     clientIP,
		DetectedVia:  detectedVia,
		IPv4Address:  ipv4,
		IPv6Address:  ipv6,
		IsPrivateIP:  isPrivateIP(clientIP),
		IsCloudflare: isCloudflareRequest(r),
		UserAgent:    r.Header.Get("User-Agent"),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
}

// ipv4Handler handles requests for IPv4 addresses only
func ipv4Handler(w http.ResponseWriter, r *http.Request) {
	ipv4 := findIPv4(r)
	
	if ipv4 == "" {
		http.Error(w, "No IPv4 address found", http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv4)
}

// ipv6Handler handles requests for IPv6 addresses only
func ipv6Handler(w http.ResponseWriter, r *http.Request) {
	ipv6 := findIPv6(r)
	
	if ipv6 == "" {
		http.Error(w, "No IPv6 address found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv6)
}

// infoHandler provides detailed IP information in plain text
func infoHandler(w http.ResponseWriter, r *http.Request) {
	info := getIPInfo(r)
	
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

// jsonHandler provides comprehensive JSON response
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	info := getIPInfo(r)
	
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

// headersHandler shows all HTTP headers and IP details for debugging
func headersHandler(w http.ResponseWriter, r *http.Request) {
	info := getIPInfo(r)
	
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

// healthHandler provides health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Set up routes
	http.HandleFunc("/", ipv4Handler)
	http.HandleFunc("/ipv6", ipv6Handler)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/headers", headersHandler)
	http.HandleFunc("/health", healthHandler)
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	addr := ":" + port
	log.Printf("Server starting on port %s", port)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

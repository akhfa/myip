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

// IPInfo contains information about the detected IP
type IPInfo struct {
	ClientIP        string            `json:"client_ip"`
	DetectedVia     string            `json:"detected_via"`
	XForwardedFor   string            `json:"x_forwarded_for,omitempty"`
	XRealIP         string            `json:"x_real_ip,omitempty"`
	CFConnectingIP  string            `json:"cf_connecting_ip,omitempty"`
	TrueClientIP    string            `json:"true_client_ip,omitempty"`
	RemoteAddr      string            `json:"remote_addr"`
	UserAgent       string            `json:"user_agent,omitempty"`
	AllHeaders      map[string]string `json:"all_headers,omitempty"`
	IsPrivateIP     bool              `json:"is_private_ip"`
	IsCloudflare    bool              `json:"is_cloudflare"`
	ProxyChain      []string          `json:"proxy_chain,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// List of headers to check for real IP (in order of priority)
var ipHeaders = []string{
	"CF-Connecting-IP",     // Cloudflare
	"True-Client-IP",       // Cloudflare Enterprise
	"X-Real-IP",           // nginx proxy/FastCGI
	"X-Forwarded-For",     // Standard proxy header
	"X-Client-IP",         // Apache mod_proxy_http
	"X-Cluster-Client-IP", // Cluster environments
	"X-Forwarded",         // Less common
	"Forwarded-For",       // Less common
	"Forwarded",           // RFC 7239
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", ipv4Handler)           // IPv4 only
	http.HandleFunc("/ipv6", ipv6Handler)       // IPv6 only
	http.HandleFunc("/info", infoHandler)       // Detailed info (old "/" response)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/headers", headersHandler)

	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  GET /       - IPv4 address only\n")
	fmt.Printf("  GET /ipv6   - IPv6 address only (404 if not available)\n")
	fmt.Printf("  GET /info   - Detailed IP information\n")
	fmt.Printf("  GET /json   - Detailed JSON response\n")
	fmt.Printf("  GET /headers - All headers display\n")
	fmt.Printf("  GET /health - Health check\n\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// IPv4 only handler
func ipv4Handler(w http.ResponseWriter, r *http.Request) {
	clientIP, _ := extractClientIP(r)
	
	// Check if it's a valid IPv4 address
	ip := net.ParseIP(clientIP)
	if ip == nil {
		http.Error(w, "Invalid IP address", http.StatusInternalServerError)
		return
	}
	
	// Check if it's IPv4
	if ip.To4() == nil {
		// It's IPv6, try to find IPv4
		ipv4 := findIPv4(r)
		if ipv4 == "" {
			http.Error(w, "No IPv4 address found", http.StatusNotFound)
			return
		}
		clientIP = ipv4
	}
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, clientIP)
}

// IPv6 only handler
func ipv6Handler(w http.ResponseWriter, r *http.Request) {
	ipv6 := findIPv6(r)
	if ipv6 == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv6)
}

// Detailed IP information handler (old "/" response)
func infoHandler(w http.ResponseWriter, r *http.Request) {
	ipInfo := getIPInfo(r)
	
	w.Header().Set("Content-Type", "text/plain")
	
	fmt.Fprintf(w, "Your IP Address: %s\n", ipInfo.ClientIP)
	fmt.Fprintf(w, "Detection Method: %s\n", ipInfo.DetectedVia)
	fmt.Fprintf(w, "Is Private IP: %t\n", ipInfo.IsPrivateIP)
	fmt.Fprintf(w, "Behind Cloudflare: %t\n", ipInfo.IsCloudflare)
	
	// Show IPv4 and IPv6 separately
	ipv4 := findIPv4(r)
	ipv6 := findIPv6(r)
	
	if ipv4 != "" {
		fmt.Fprintf(w, "IPv4 Address: %s\n", ipv4)
	}
	if ipv6 != "" {
		fmt.Fprintf(w, "IPv6 Address: %s\n", ipv6)
	}
	
	if len(ipInfo.ProxyChain) > 0 {
		fmt.Fprintf(w, "Proxy Chain: %s\n", strings.Join(ipInfo.ProxyChain, " → "))
	}
	
	fmt.Fprintf(w, "Timestamp: %s\n", ipInfo.Timestamp.Format(time.RFC3339))
}

// JSON response handler
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	ipInfo := getIPInfo(r)
	
	// Add IPv4 and IPv6 specific fields
	enrichedInfo := struct {
		*IPInfo
		IPv4Address string `json:"ipv4_address,omitempty"`
		IPv6Address string `json:"ipv6_address,omitempty"`
	}{
		IPInfo:      ipInfo,
		IPv4Address: findIPv4(r),
		IPv6Address: findIPv6(r),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if err := json.NewEncoder(w).Encode(enrichedInfo); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// Headers display handler
func headersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	
	fmt.Fprintf(w, "All HTTP Headers:\n")
	fmt.Fprintf(w, "=================\n\n")
	
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", name, value)
		}
	}
	
	fmt.Fprintf(w, "\nRemote Address: %s\n", r.RemoteAddr)
	
	ipInfo := getIPInfo(r)
	fmt.Fprintf(w, "\nDetected IP: %s (via %s)\n", ipInfo.ClientIP, ipInfo.DetectedVia)
	
	ipv4 := findIPv4(r)
	ipv6 := findIPv6(r)
	if ipv4 != "" {
		fmt.Fprintf(w, "IPv4: %s\n", ipv4)
	}
	if ipv6 != "" {
		fmt.Fprintf(w, "IPv6: %s\n", ipv6)
	}
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Find IPv4 address from request
func findIPv4(r *http.Request) string {
	// Check all possible IP sources
	allIPs := getAllPossibleIPs(r)
	
	for _, ip := range allIPs {
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil && parsedIP.To4() != nil {
			return ip
		}
	}
	
	return ""
}

// Find IPv6 address from request
func findIPv6(r *http.Request) string {
	// Check all possible IP sources
	allIPs := getAllPossibleIPs(r)
	
	for _, ip := range allIPs {
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil && parsedIP.To4() == nil {
			// It's IPv6, but let's clean it up (remove brackets if present)
			cleanIP := strings.Trim(ip, "[]")
			if net.ParseIP(cleanIP) != nil {
				return cleanIP
			}
		}
	}
	
	return ""
}

// Get all possible IPs from various headers and sources
func getAllPossibleIPs(r *http.Request) []string {
	var ips []string
	
	// Check all headers
	for _, header := range ipHeaders {
		value := r.Header.Get(header)
		if value != "" {
			// Handle comma-separated IPs
			headerIPs := strings.Split(value, ",")
			for _, ip := range headerIPs {
				cleanIP := strings.TrimSpace(ip)
				if isValidIP(cleanIP) {
					ips = append(ips, cleanIP)
				}
			}
		}
	}
	
	// Add RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if isValidIP(host) {
			ips = append(ips, host)
		}
	} else if isValidIP(r.RemoteAddr) {
		ips = append(ips, r.RemoteAddr)
	}
	
	// Remove duplicates
	return removeDuplicates(ips)
}

// Remove duplicate IPs
func removeDuplicates(ips []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, ip := range ips {
		if !keys[ip] {
			keys[ip] = true
			result = append(result, ip)
		}
	}
	
	return result
}

// Main function to extract IP information
func getIPInfo(r *http.Request) *IPInfo {
	info := &IPInfo{
		RemoteAddr:  r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		AllHeaders:  make(map[string]string),
		Timestamp:   time.Now(),
		IsCloudflare: isCloudflareRequest(r),
	}

	// Copy all headers
	for name, values := range r.Header {
		if len(values) > 0 {
			info.AllHeaders[name] = values[0]
		}
	}

	// Extract specific header values
	info.XForwardedFor = r.Header.Get("X-Forwarded-For")
	info.XRealIP = r.Header.Get("X-Real-IP")
	info.CFConnectingIP = r.Header.Get("CF-Connecting-IP")
	info.TrueClientIP = r.Header.Get("True-Client-IP")

	// Determine the real client IP
	clientIP, method := extractClientIP(r)
	info.ClientIP = clientIP
	info.DetectedVia = method
	info.IsPrivateIP = isPrivateIP(clientIP)

	// Build proxy chain
	info.ProxyChain = buildProxyChain(r)

	return info
}

// Extract the real client IP address
func extractClientIP(r *http.Request) (string, string) {
	// Check each header in order of priority
	for _, header := range ipHeaders {
		value := r.Header.Get(header)
		if value != "" {
			// Handle comma-separated IPs (especially for X-Forwarded-For)
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				cleanIP := strings.TrimSpace(ip)
				if isValidIP(cleanIP) && !isPrivateIP(cleanIP) {
					return cleanIP, header
				}
			}
			// If all IPs are private, return the first valid one
			for _, ip := range ips {
				cleanIP := strings.TrimSpace(ip)
				if isValidIP(cleanIP) {
					return cleanIP, header + " (private)"
				}
			}
		}
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, "RemoteAddr (raw)"
	}
	return host, "RemoteAddr"
}

// Build the proxy chain from headers
func buildProxyChain(r *http.Request) []string {
	var chain []string

	// Check X-Forwarded-For for proxy chain
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			cleanIP := strings.TrimSpace(ip)
			if isValidIP(cleanIP) {
				chain = append(chain, cleanIP)
			}
		}
	}

	// Add RemoteAddr if not already in chain
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if len(chain) == 0 || chain[len(chain)-1] != host {
			chain = append(chain, host)
		}
	}

	return chain
}

// Check if the request is coming through Cloudflare
func isCloudflareRequest(r *http.Request) bool {
	// Check for Cloudflare-specific headers
	cfHeaders := []string{
		"CF-Connecting-IP",
		"CF-Ray",
		"CF-Visitor",
		"CF-IPCountry",
	}

	for _, header := range cfHeaders {
		if r.Header.Get(header) != "" {
			return true
		}
	}

	return false
}

// Validate if string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// Check if IP is private/internal
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Define private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 private
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}
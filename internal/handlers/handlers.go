package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"myip/internal/ip"
	"myip/internal/models"
)

// isJSONFormat checks if format parameter equals "json" case-insensitively
// Optimized for performance - avoids string allocation from ToLower()
func isJSONFormat(format string) bool {
	if len(format) != 4 {
		return false
	}
	// Check each byte directly for maximum performance
	return (format[0] == 'j' || format[0] == 'J') &&
		(format[1] == 's' || format[1] == 'S') &&
		(format[2] == 'o' || format[2] == 'O') &&
		(format[3] == 'n' || format[3] == 'N')
}

// isJSONPFormat checks if format parameter equals "jsonp" case-insensitively
// Optimized for performance - avoids string allocation from ToLower()
func isJSONPFormat(format string) bool {
	if len(format) != 5 {
		return false
	}
	// Check each byte directly for maximum performance
	return (format[0] == 'j' || format[0] == 'J') &&
		(format[1] == 's' || format[1] == 'S') &&
		(format[2] == 'o' || format[2] == 'O') &&
		(format[3] == 'n' || format[3] == 'N') &&
		(format[4] == 'p' || format[4] == 'P')
}

// validCallbackRegex matches valid JavaScript identifier names for JSONP callbacks
// Allows letters, digits, underscore, and dot notation (for object methods)
var validCallbackRegex = regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$.]*$`)

// sanitizeCallback ensures the callback parameter contains only valid JavaScript identifier characters
// This prevents XSS attacks via callback parameter injection
func sanitizeCallback(callback string) string {
	if callback == "" {
		return "callback"
	}
	
	// Limit callback length to prevent abuse
	if len(callback) > 50 {
		return "callback"
	}
	
	// Validate callback contains only safe JavaScript identifier characters
	if !validCallbackRegex.MatchString(callback) {
		return "callback"
	}
	
	return callback
}

// IPv4Handler handles requests for IPv4 addresses only
// @Summary Get IPv4 address
// @Description Returns the client's IPv4 address in plain text format, JSON format if format=json, or JSONP format if format=jsonp is specified (case-insensitive). Callback parameter only works with format=jsonp.
// @Tags IP Detection
// @Accept json
// @Produce plain,json
// @Param format query string false "Response format (json for JSON response, jsonp for JSONP response)"
// @Param callback query string false "Callback function name for JSONP response. Only works with format=jsonp. Without format=jsonp, callback parameter is ignored and returns plain text (ipify.org compatible behavior). (default: callback)"
// @Success 200 {string} string "IPv4 address (plain text)"
// @Success 200 {object} map[string]string "IP address in JSON format: {\"ip\": \"192.168.1.1\"}"
// @Success 200 {string} string "IP address in JSONP format: callback({\"ip\": \"192.168.1.1\"}) or getip({\"ip\": \"192.168.1.1\"}) with custom callback"
// @Failure 404 {string} string "No IPv4 address found"
// @Router / [get]
func IPv4Handler(w http.ResponseWriter, r *http.Request) {
	ipv4 := ip.FindIPv4(r)

	if ipv4 == "" {
		http.Error(w, "No IPv4 address found", http.StatusNotFound)
		return
	}

	// Check format parameter
	format := r.URL.Query().Get("format")
	callback := r.URL.Query().Get("callback")

	// Check if JSONP format is requested (case-insensitive, optimized)
	if isJSONPFormat(format) {
		sanitizedCallback := sanitizeCallback(callback)
		
		w.Header().Set("Content-Type", "application/javascript")
		
		// Use proper JSON encoding to prevent injection attacks
		response := map[string]string{"ip": ipv4}
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to encode JSONP response for IPv4 %s: %v", ipv4, err)
			http.Error(w, "Failed to encode JSONP response", http.StatusInternalServerError)
			return
		}
		
		fmt.Fprintf(w, "%s(%s);", sanitizedCallback, string(jsonBytes))
		return
	}

	// Check if JSON format is requested (case-insensitive, optimized)
	if isJSONFormat(format) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"ip": ipv4}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode JSON response for IPv4 %s: %v", ipv4, err)
			http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Default plain text response
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv4)
}

// IPv6Handler handles requests for IPv6 addresses only
// @Summary Get IPv6 address
// @Description Returns the client's IPv6 address in plain text format, JSON format if format=json, or JSONP format if format=jsonp is specified (case-insensitive). Callback parameter only works with format=jsonp.
// @Tags IP Detection
// @Accept json
// @Produce plain,json
// @Param format query string false "Response format (json for JSON response, jsonp for JSONP response)"
// @Param callback query string false "Callback function name for JSONP response. Only works with format=jsonp. Without format=jsonp, callback parameter is ignored and returns plain text (ipify.org compatible behavior). (default: callback)"
// @Success 200 {string} string "IPv6 address (plain text)"
// @Success 200 {object} map[string]string "IP address in JSON format: {\"ip\": \"2001:db8::1\"}"
// @Success 200 {string} string "IP address in JSONP format: callback({\"ip\": \"2001:db8::1\"}) or getip({\"ip\": \"2001:db8::1\"}) with custom callback"
// @Failure 404 {string} string "No IPv6 address found"
// @Router /ipv6 [get]
func IPv6Handler(w http.ResponseWriter, r *http.Request) {
	ipv6 := ip.FindIPv6(r)

	if ipv6 == "" {
		http.Error(w, "No IPv6 address found", http.StatusNotFound)
		return
	}

	// Check format parameter
	format := r.URL.Query().Get("format")
	callback := r.URL.Query().Get("callback")

	// Check if JSONP format is requested (case-insensitive, optimized)
	if isJSONPFormat(format) {
		sanitizedCallback := sanitizeCallback(callback)
		
		w.Header().Set("Content-Type", "application/javascript")
		
		// Use proper JSON encoding to prevent injection attacks
		response := map[string]string{"ip": ipv6}
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to encode JSONP response for IPv6 %s: %v", ipv6, err)
			http.Error(w, "Failed to encode JSONP response", http.StatusInternalServerError)
			return
		}
		
		fmt.Fprintf(w, "%s(%s);", sanitizedCallback, string(jsonBytes))
		return
	}

	// Check if JSON format is requested (case-insensitive, optimized)
	if isJSONFormat(format) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"ip": ipv6}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode JSON response for IPv6 %s: %v", ipv6, err)
			http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Default plain text response
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, ipv6)
}

// InfoHandler provides detailed IP information in plain text
// @Summary Get detailed IP information
// @Description Returns comprehensive IP information including detection method, private IP status, and Cloudflare detection in plain text format
// @Tags IP Detection
// @Accept json
// @Produce plain
// @Success 200 {string} string "Detailed IP information in plain text"
// @Router /info [get]
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
// @Summary Get IP information in JSON format
// @Description Returns comprehensive IP information in JSON format including all detected addresses, detection method, and metadata
// @Tags IP Detection
// @Accept json
// @Produce json
// @Success 200 {object} models.IPInfo "IP information in JSON format"
// @Failure 500 {string} string "Failed to encode JSON response"
// @Router /json [get]
func JSONHandler(w http.ResponseWriter, r *http.Request) {
	info := ip.GetInfo(r)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

// HeadersHandler shows all HTTP headers and IP details for debugging
// @Summary Debug headers and connection information
// @Description Returns all HTTP headers, IP detection details, and connection information for debugging purposes
// @Tags Debug
// @Accept json
// @Produce plain
// @Success 200 {string} string "Complete debugging information including headers and connection details"
// @Router /headers [get]
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
// @Summary Health check
// @Description Returns service health status and timestamp
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse "Service health status"
// @Failure 500 {string} string "Failed to encode health response"
// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := models.NewHealthResponse("healthy")

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		return
	}
}

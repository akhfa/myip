package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"myip/internal/models"
)

func TestIPv4Handler(t *testing.T) {
	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedCode int
	}{
		{
			name: "Valid IPv4",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusOK,
		},
		{
			name:         "IPv4 from RemoteAddr",
			headers:      map[string]string{},
			remoteAddr:   "203.0.113.1:12345",
			expectedCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(IPv4Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, test.expectedCode)
			}
		})
	}
}

func TestIPv6Handler(t *testing.T) {
	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedCode int
		expectBody   bool
	}{
		{
			name: "Valid IPv6",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			remoteAddr:   "[::1]:12345",
			expectedCode: http.StatusOK,
			expectBody:   true,
		},
		{
			name:         "IPv6 from RemoteAddr",
			headers:      map[string]string{},
			remoteAddr:   "[2001:db8::1]:12345",
			expectedCode: http.StatusOK,
			expectBody:   true,
		},
		{
			name: "No IPv6 found - only IPv4",
			headers: map[string]string{
				"CF-Connecting-IP": "192.168.1.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusNotFound,
			expectBody:   false,
		},
		{
			name:         "No IPv6 found - empty headers",
			headers:      map[string]string{},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusNotFound,
			expectBody:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ipv6", nil)
			req.RemoteAddr = test.remoteAddr

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(IPv6Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, test.expectedCode)
			}

			if test.expectBody && strings.TrimSpace(rr.Body.String()) == "" {
				t.Errorf("expected body content but got empty response")
			}
		})
	}
}

func TestInfoHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/info", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(InfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedStrings := []string{
		"Your IP Address: 203.0.113.1",
		"Detection Method: CF-Connecting-IP",
		"Is Private IP: false",
		"Behind Cloudflare: true",
		"IPv4 Address: 203.0.113.1",
		"Timestamp:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected body to contain %q, but it didn't. Body: %s", expected, body)
		}
	}

	// Test content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", contentType)
	}
}

func TestJSONHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/json", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(JSONHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Test content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Test JSON structure
	var response models.IPInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.ClientIP != "203.0.113.1" {
		t.Errorf("Expected ClientIP 203.0.113.1, got %s", response.ClientIP)
	}

	if response.DetectedVia != "CF-Connecting-IP" {
		t.Errorf("Expected DetectedVia CF-Connecting-IP, got %s", response.DetectedVia)
	}

	if !response.IsCloudflare {
		t.Error("Expected IsCloudflare to be true")
	}

	if response.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected UserAgent TestAgent/1.0, got %s", response.UserAgent)
	}
}

func TestHeadersHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Custom-Header", "test-value")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HeadersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedStrings := []string{
		"=== IP INFORMATION ===",
		"Client IP: 203.0.113.1",
		"Detection Method: CF-Connecting-IP",
		"=== HTTP HEADERS ===",
		"Cf-Connecting-Ip: 203.0.113.1",
		"User-Agent: TestAgent/1.0",
		"X-Custom-Header: test-value",
		"=== CONNECTION INFO ===",
		"Remote Address: 192.168.1.1:12345",
		"Method: GET",
		"URL: /headers",
		"Protocol: HTTP/1.1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected body to contain %q, but it didn't. Body: %s", expected, body)
		}
	}

	// Test content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", contentType)
	}
}

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "healthy"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Test JSON structure
	var response models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", response.Status)
	}

	if response.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
}

// Additional tests for better coverage

func TestIPv4HandlerErrorCases(t *testing.T) {
	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedCode int
	}{
		{
			name:         "No IPv4 found - only IPv6",
			headers:      map[string]string{"X-Forwarded-For": "2001:db8::1"},
			remoteAddr:   "[2001:db8::1]:12345",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "No valid IP found",
			headers:      map[string]string{"X-Forwarded-For": "invalid"},
			remoteAddr:   "invalid:12345",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(IPv4Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, test.expectedCode)
			}
		})
	}
}

func TestJSONHandlerErrorCase(t *testing.T) {
	// This test is more for completeness - the actual JSON encoding rarely fails
	// in normal circumstances, but we test the successful path
	req := httptest.NewRequest("GET", "/json", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(JSONHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Verify it's valid JSON
	var response models.IPInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
}

func TestHealthHandlerErrorCase(t *testing.T) {
	// Similar to JSON handler - testing the successful path
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Verify it's valid JSON
	var response models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
}

func TestInfoHandlerEmptyIPv6(t *testing.T) {
	// Test info handler when IPv6 is empty
	req := httptest.NewRequest("GET", "/info", nil)
	req.Header.Set("CF-Connecting-IP", "192.168.1.1") // IPv4 only
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "10.0.0.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(InfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()

	// Should contain IPv4 info but not IPv6
	if !strings.Contains(body, "IPv4 Address: 192.168.1.1") {
		t.Error("Expected IPv4 address in response")
	}

	// Should not contain IPv6 line since it's empty
	if strings.Contains(body, "IPv6 Address:") {
		t.Error("Should not contain IPv6 address line when IPv6 is empty")
	}
}

func TestInfoHandlerEmptyIPv4(t *testing.T) {
	// Test info handler when IPv4 is empty
	req := httptest.NewRequest("GET", "/info", nil)
	req.Header.Set("CF-Connecting-IP", "2001:db8::1") // IPv6 only
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "[::1]:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(InfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()

	// Should contain IPv6 info but not IPv4
	if !strings.Contains(body, "IPv6 Address: 2001:db8::1") {
		t.Error("Expected IPv6 address in response")
	}

	// Should not contain IPv4 line since it's empty
	if strings.Contains(body, "IPv4 Address:") {
		t.Error("Should not contain IPv4 address line when IPv4 is empty")
	}
}

func TestHeadersHandlerEmptyAddresses(t *testing.T) {
	// Test headers handler with minimal IP info
	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "invalid-addr" // This should fallback

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HeadersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	expectedStrings := []string{
		"=== IP INFORMATION ===",
		"Client IP: invalid-addr",
		"Detection Method: RemoteAddr",
		"=== HTTP HEADERS ===",
		"User-Agent: TestAgent/1.0",
		"=== CONNECTION INFO ===",
		"Remote Address: invalid-addr",
		"Method: GET",
		"URL: /headers",
		"Protocol: HTTP/1.1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected body to contain %q, but it didn't. Body: %s", expected, body)
		}
	}
}

// TestIPv4HandlerJSONFormat tests the new format=json query parameter functionality
func TestIPv4HandlerJSONFormat(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		headers      map[string]string
		remoteAddr   string
		expectedCode int
		expectJSON   bool
	}{
		{
			name:  "JSON format requested",
			query: "?format=json",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:  "Plain text format (default)",
			query: "",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusOK,
			expectJSON:   false,
		},
		{
			name:  "JSON format requested - uppercase",
			query: "?format=JSON",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:  "JSON format requested - mixed case",
			query: "?format=Json",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:  "Other format parameter ignored",
			query: "?format=xml",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr:   "192.168.1.1:12345",
			expectedCode: http.StatusOK,
			expectJSON:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+test.query, nil)
			req.RemoteAddr = test.remoteAddr

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(IPv4Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.expectedCode)
			}

			if test.expectJSON {
				// Check content type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected content type application/json, got %s", contentType)
				}

				// Parse and validate JSON response
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				// Verify JSON contains expected IPv4
				if response["ip"] != "203.0.113.1" {
					t.Errorf("Expected ip to be 203.0.113.1, got %s", response["ip"])
				}
			} else {
				// Check content type for plain text
				contentType := rr.Header().Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("Expected content type text/plain, got %s", contentType)
				}

				// Check plain text response
				body := rr.Body.String()
				if !strings.Contains(body, "203.0.113.1") {
					t.Errorf("Expected body to contain 203.0.113.1, got %s", body)
				}
			}
		})
	}
}

// TestIPv6HandlerJSONFormat tests the new format=json query parameter functionality for IPv6
func TestIPv6HandlerJSONFormat(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		headers      map[string]string
		remoteAddr   string
		expectedCode int
		expectJSON   bool
	}{
		{
			name:  "JSON format requested with IPv6",
			query: "?format=json",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			remoteAddr:   "[2001:db8::1]:12345",
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:  "Plain text format with IPv6 (default)",
			query: "",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			remoteAddr:   "[2001:db8::1]:12345",
			expectedCode: http.StatusOK,
			expectJSON:   false,
		},
		{
			name:  "JSON format requested with IPv6 - uppercase",
			query: "?format=JSON",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			remoteAddr:   "[2001:db8::1]:12345",
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:  "JSON format requested with IPv6 - mixed case",
			query: "?format=Json",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			remoteAddr:   "[2001:db8::1]:12345",
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ipv6"+test.query, nil)
			req.RemoteAddr = test.remoteAddr

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(IPv6Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.expectedCode)
			}

			if test.expectJSON {
				// Check content type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected content type application/json, got %s", contentType)
				}

				// Parse and validate JSON response
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				// Verify JSON contains expected IPv6
				if response["ip"] != "2001:db8::1" {
					t.Errorf("Expected ip to be 2001:db8::1, got %s", response["ip"])
				}
			} else {
				// Check content type for plain text
				contentType := rr.Header().Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("Expected content type text/plain, got %s", contentType)
				}

				// Check plain text response
				body := rr.Body.String()
				if !strings.Contains(body, "2001:db8::1") {
					t.Errorf("Expected body to contain 2001:db8::1, got %s", body)
				}
			}
		})
	}
}

// Test the isJSONFormat function thoroughly
func TestIsJSONFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"json", true},
		{"JSON", true},
		{"Json", true},
		{"jSoN", true},
		{"JsOn", true},
		{"xml", false},
		{"text", false},
		{"", false},
		{"jsonformat", false}, // too long
		{"jso", false},        // too short
		{"html", false},
		{"yaml", false},
		{"j", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := isJSONFormat(test.input)
			if result != test.expected {
				t.Errorf("isJSONFormat(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

// failingWriter is a custom ResponseWriter that fails on Write operations
type failingWriter struct {
	header     http.Header
	written    bool
	statusCode int
}

func (fw *failingWriter) Header() http.Header {
	if fw.header == nil {
		fw.header = make(http.Header)
	}
	return fw.header
}

func (fw *failingWriter) Write(data []byte) (int, error) {
	fw.written = true
	// Check if this is the JSON response attempt (not the error response)
	if !strings.Contains(string(data), "Failed to encode JSON response") {
		return 0, fmt.Errorf("write failed")
	}
	// Let error responses through
	return len(data), nil
}

func (fw *failingWriter) WriteHeader(code int) {
	fw.statusCode = code
}

// TestJSONEncodingErrors tests JSON encoding error paths in handlers
func TestJSONEncodingErrors(t *testing.T) {

	t.Run("IPv4Handler JSON encoding error", func(t *testing.T) {
		// Create request with format=json
		req := httptest.NewRequest("GET", "/?format=json", nil)
		req.Header.Set("X-Real-IP", "192.168.1.100")

		fw := &failingWriter{}

		IPv4Handler(fw, req)

		if !fw.written {
			t.Error("Expected write to be attempted")
		}

		// Check that the error response status was set
		if fw.statusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, fw.statusCode)
		}

		// Check that the error content type was set by http.Error
		if fw.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
			t.Errorf("Expected Content-Type to be text/plain; charset=utf-8, got %s", fw.Header().Get("Content-Type"))
		}
	})

	t.Run("IPv6Handler JSON encoding error", func(t *testing.T) {
		// Create request with format=json
		req := httptest.NewRequest("GET", "/?format=json", nil)
		req.Header.Set("X-Real-IP", "2001:db8::1")

		fw := &failingWriter{}

		IPv6Handler(fw, req)

		if !fw.written {
			t.Error("Expected write to be attempted")
		}

		// Check that the error response status was set
		if fw.statusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, fw.statusCode)
		}
	})

	t.Run("JSONHandler encoding error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "192.168.1.100")

		fw := &failingWriter{}

		JSONHandler(fw, req)

		if !fw.written {
			t.Error("Expected write to be attempted")
		}

		// Check that the error response status was set
		if fw.statusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, fw.statusCode)
		}
	})

	t.Run("HealthHandler encoding error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		fw := &failingWriter{}

		HealthHandler(fw, req)

		if !fw.written {
			t.Error("Expected write to be attempted")
		}

		// Check that the error response status was set
		if fw.statusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, fw.statusCode)
		}
	})
}

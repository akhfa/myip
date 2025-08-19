package handlers

import (
	"encoding/json"
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
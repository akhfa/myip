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

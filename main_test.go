package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
	"myip/internal/config"
	"myip/internal/handlers"
)

// Integration tests for the main application endpoints
func TestIntegrationHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.HealthHandler)

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
}

func TestIntegrationIPv4Handler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.IPv4Handler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "203.0.113.1"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestIntegrationIPv6Handler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ipv6", nil)
	req.Header.Set("CF-Connecting-IP", "2001:db8::1")
	req.RemoteAddr = "[::1]:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.IPv6Handler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "2001:db8::1"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestIntegrationInfoHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/info", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.InfoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Your IP Address: 203.0.113.1") {
		t.Error("Expected IP address in info response")
	}

	if !strings.Contains(body, "Behind Cloudflare: true") {
		t.Error("Expected Cloudflare detection in info response")
	}
}

func TestIntegrationJSONHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/json", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.JSONHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Check JSON structure
	if !strings.Contains(rr.Body.String(), "client_ip") {
		t.Error("Expected JSON to contain client_ip field")
	}
}

func TestIntegrationHeadersHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.HeadersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "=== IP INFORMATION ===") {
		t.Error("Expected IP information section in headers response")
	}

	if !strings.Contains(body, "=== HTTP HEADERS ===") {
		t.Error("Expected HTTP headers section in headers response")
	}
}

// Test the main function coverage indirectly by testing configuration loading
func TestConfigurationIntegration(t *testing.T) {
	// Test default configuration
	cfg := config.Load()
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}
	
	if cfg.GetAddr() != ":8080" {
		t.Errorf("Expected address :8080, got %s", cfg.GetAddr())
	}
}

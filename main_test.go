package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
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

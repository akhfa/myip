package config

import (
	"os"
	"testing"
)

func TestLoadDefaultPort(t *testing.T) {
	// Ensure no PORT environment variable is set
	os.Unsetenv("PORT")
	os.Unsetenv("HOST")
	
	cfg := Load()
	
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}
	
	if cfg.Host != "localhost:8080" {
		t.Errorf("Expected default host localhost:8080, got %s", cfg.Host)
	}
	
	expectedAddr := ":8080"
	if cfg.GetAddr() != expectedAddr {
		t.Errorf("Expected address %s, got %s", expectedAddr, cfg.GetAddr())
	}
}

func TestLoadCustomPort(t *testing.T) {
	// Set custom PORT environment variable
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")
	
	cfg := Load()
	
	if cfg.Port != "3000" {
		t.Errorf("Expected custom port 3000, got %s", cfg.Port)
	}
	
	expectedAddr := ":3000"
	if cfg.GetAddr() != expectedAddr {
		t.Errorf("Expected address %s, got %s", expectedAddr, cfg.GetAddr())
	}
}

func TestLoadEmptyPortFallback(t *testing.T) {
	// Set empty PORT environment variable
	os.Setenv("PORT", "")
	defer os.Unsetenv("PORT")
	
	cfg := Load()
	
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080 when PORT is empty, got %s", cfg.Port)
	}
	
	expectedAddr := ":8080"
	if cfg.GetAddr() != expectedAddr {
		t.Errorf("Expected address %s, got %s", expectedAddr, cfg.GetAddr())
	}
}

func TestGetAddrFormat(t *testing.T) {
	tests := []struct {
		port         string
		expectedAddr string
	}{
		{"8080", ":8080"},
		{"3000", ":3000"},
		{"80", ":80"},
		{"9999", ":9999"},
	}
	
	for _, test := range tests {
		cfg := &Config{Port: test.port}
		result := cfg.GetAddr()
		
		if result != test.expectedAddr {
			t.Errorf("For port %s, expected %s, got %s", test.port, test.expectedAddr, result)
		}
	}
}

func TestLoadDefaultHost(t *testing.T) {
	// Ensure no HOST environment variable is set
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")
	
	cfg := Load()
	
	if cfg.Host != "localhost:8080" {
		t.Errorf("Expected default host localhost:8080, got %s", cfg.Host)
	}
}

func TestLoadCustomHost(t *testing.T) {
	// Set custom HOST environment variable
	os.Setenv("HOST", "example.com")
	defer os.Unsetenv("HOST")
	
	cfg := Load()
	
	if cfg.Host != "example.com" {
		t.Errorf("Expected custom host example.com, got %s", cfg.Host)
	}
}

func TestLoadEmptyHostFallback(t *testing.T) {
	// Set empty HOST environment variable
	os.Setenv("HOST", "")
	defer os.Unsetenv("HOST")
	
	cfg := Load()
	
	if cfg.Host != "localhost:8080" {
		t.Errorf("Expected default host localhost:8080 when HOST is empty, got %s", cfg.Host)
	}
}
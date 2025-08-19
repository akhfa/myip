package models

import (
	"testing"
	"time"
)

func TestNewHealthResponse(t *testing.T) {
	status := "healthy"
	response := NewHealthResponse(status)
	
	if response.Status != status {
		t.Errorf("Expected status %s, got %s", status, response.Status)
	}
	
	if response.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
	
	// Verify timestamp is in RFC3339 format
	_, err := time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		t.Errorf("Timestamp is not in RFC3339 format: %v", err)
	}
}

func TestNewHealthResponseCustomStatus(t *testing.T) {
	testCases := []string{
		"unhealthy",
		"degraded",
		"maintenance",
		"ok",
	}
	
	for _, status := range testCases {
		response := NewHealthResponse(status)
		
		if response.Status != status {
			t.Errorf("Expected status %s, got %s", status, response.Status)
		}
		
		if response.Timestamp == "" {
			t.Errorf("Expected timestamp to be set for status %s", status)
		}
	}
}

func TestNewHealthResponseTimestamp(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	response := NewHealthResponse("healthy")
	after := time.Now().UTC().Add(time.Second)
	
	timestamp, err := time.Parse(time.RFC3339, response.Timestamp)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	
	// Verify timestamp is between before and after (with 1 second buffer)
	if timestamp.Before(before) || timestamp.After(after) {
		t.Errorf("Timestamp %v is not between %v and %v", timestamp, before, after)
	}
}

func TestIPInfoStruct(t *testing.T) {
	// Test that we can create and populate an IPInfo struct
	info := &IPInfo{
		ClientIP:      "203.0.113.1",
		DetectedVia:   "CF-Connecting-IP",
		IPv4Address:   "203.0.113.1",
		IPv6Address:   "2001:db8::1",
		IsPrivateIP:   false,
		IsCloudflare:  true,
		UserAgent:     "TestAgent/1.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
	
	if info.ClientIP != "203.0.113.1" {
		t.Errorf("Expected ClientIP 203.0.113.1, got %s", info.ClientIP)
	}
	
	if info.DetectedVia != "CF-Connecting-IP" {
		t.Errorf("Expected DetectedVia CF-Connecting-IP, got %s", info.DetectedVia)
	}
	
	if info.IPv4Address != "203.0.113.1" {
		t.Errorf("Expected IPv4Address 203.0.113.1, got %s", info.IPv4Address)
	}
	
	if info.IPv6Address != "2001:db8::1" {
		t.Errorf("Expected IPv6Address 2001:db8::1, got %s", info.IPv6Address)
	}
	
	if info.IsPrivateIP != false {
		t.Errorf("Expected IsPrivateIP false, got %t", info.IsPrivateIP)
	}
	
	if info.IsCloudflare != true {
		t.Errorf("Expected IsCloudflare true, got %t", info.IsCloudflare)
	}
	
	if info.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected UserAgent TestAgent/1.0, got %s", info.UserAgent)
	}
	
	if info.Timestamp == "" {
		t.Error("Expected Timestamp to be set")
	}
}

func TestHealthResponseStruct(t *testing.T) {
	// Test that we can create and populate a HealthResponse struct
	timestamp := time.Now().UTC().Format(time.RFC3339)
	response := &HealthResponse{
		Status:    "healthy",
		Timestamp: timestamp,
	}
	
	if response.Status != "healthy" {
		t.Errorf("Expected Status healthy, got %s", response.Status)
	}
	
	if response.Timestamp != timestamp {
		t.Errorf("Expected Timestamp %s, got %s", timestamp, response.Timestamp)
	}
}
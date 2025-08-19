package ip

import (
	"net/http/httptest"
	"testing"
)

func TestGetInfo(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("CF-Ray", "123456789-ABC")
	req.RemoteAddr = "192.168.1.1:12345"

	info := GetInfo(req)

	if info.ClientIP != "203.0.113.1" {
		t.Errorf("Expected ClientIP 203.0.113.1, got %s", info.ClientIP)
	}

	if info.DetectedVia != "CF-Connecting-IP" {
		t.Errorf("Expected DetectedVia CF-Connecting-IP, got %s", info.DetectedVia)
	}

	if info.IsPrivateIP {
		t.Error("Expected IsPrivateIP to be false for public IP")
	}

	if !info.IsCloudflare {
		t.Error("Expected IsCloudflare to be true")
	}

	if info.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected UserAgent TestAgent/1.0, got %s", info.UserAgent)
	}

	if info.IPv4Address != "203.0.113.1" {
		t.Errorf("Expected IPv4Address 203.0.113.1, got %s", info.IPv4Address)
	}

	if info.Timestamp == "" {
		t.Error("Expected Timestamp to be set")
	}
}

func TestGetInfoPrivateIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "10.0.0.1:12345"

	info := GetInfo(req)

	if info.ClientIP != "192.168.1.100" {
		t.Errorf("Expected ClientIP 192.168.1.100, got %s", info.ClientIP)
	}

	if info.DetectedVia != "X-Real-IP" {
		t.Errorf("Expected DetectedVia X-Real-IP, got %s", info.DetectedVia)
	}

	if !info.IsPrivateIP {
		t.Error("Expected IsPrivateIP to be true for private IP")
	}

	if info.IsCloudflare {
		t.Error("Expected IsCloudflare to be false")
	}

	if info.IPv4Address != "192.168.1.100" {
		t.Errorf("Expected IPv4Address 192.168.1.100, got %s", info.IPv4Address)
	}
}

func TestGetInfoIPv6(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "2001:db8::1, 203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	info := GetInfo(req)

	if info.ClientIP != "2001:db8::1" {
		t.Errorf("Expected ClientIP 2001:db8::1, got %s", info.ClientIP)
	}

	if info.DetectedVia != "X-Forwarded-For" {
		t.Errorf("Expected DetectedVia X-Forwarded-For, got %s", info.DetectedVia)
	}

	if info.IPv4Address != "203.0.113.1" {
		t.Errorf("Expected IPv4Address 203.0.113.1, got %s", info.IPv4Address)
	}

	if info.IPv6Address != "2001:db8::1" {
		t.Errorf("Expected IPv6Address 2001:db8::1, got %s", info.IPv6Address)
	}
}
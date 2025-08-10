package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

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

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"invalid", false},
		{"", false},
		{"999.999.999.999", false},
	}

	for _, test := range tests {
		result := isValidIP(test.ip)
		if result != test.expected {
			t.Errorf("isValidIP(%s) = %v; want %v", test.ip, result, test.expected)
		}
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"127.0.0.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"::1", true},
		{"invalid", false},
	}

	for _, test := range tests {
		result := isPrivateIP(test.ip)
		if result != test.expected {
			t.Errorf("isPrivateIP(%s) = %v; want %v", test.ip, result, test.expected)
		}
	}
}

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "CF-Connecting-IP",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 192.168.1.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.1",
			},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "203.0.113.1:12345",
			expectedIP: "203.0.113.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr
			
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			ip, _ := extractClientIP(req)
			if ip != test.expectedIP {
				t.Errorf("extractClientIP() = %v; want %v", ip, test.expectedIP)
			}
		})
	}
}

func TestFindIPv4(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "IPv4 in CF-Connecting-IP",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.1",
		},
		{
			name: "IPv4 in X-Forwarded-For with IPv6",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1, 203.0.113.1",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.1",
		},
		{
			name:       "IPv4 in RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "203.0.113.1:12345",
			expected:   "203.0.113.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr
			
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			result := findIPv4(req)
			if result != test.expected {
				t.Errorf("findIPv4() = %v; want %v", result, test.expected)
			}
		})
	}
}

func TestFindIPv6(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "IPv6 in CF-Connecting-IP",
			headers: map[string]string{
				"CF-Connecting-IP": "2001:db8::1",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "2001:db8::1",
		},
		{
			name: "IPv6 in X-Forwarded-For with IPv4",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 2001:db8::1",
			},
			remoteAddr: "192.168.1.1:12345",
			expected:   "2001:db8::1",
		},
		{
			name:       "IPv6 in RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "[2001:db8::1]:12345",
			expected:   "2001:db8::1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr
			
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			result := findIPv6(req)
			if result != test.expected {
				t.Errorf("findIPv6() = %v; want %v", result, test.expected)
			}
		})
	}
}

func TestIsCloudflareRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{
			name: "CF-Connecting-IP present",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			expected: true,
		},
		{
			name: "CF-Ray present",
			headers: map[string]string{
				"CF-Ray": "123456789-ABC",
			},
			expected: true,
		},
		{
			name: "No Cloudflare headers",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			expected: false,
		},
		{
			name:     "No headers",
			headers:  map[string]string{},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			result := isCloudflareRequest(req)
			if result != test.expected {
				t.Errorf("isCloudflareRequest() = %v; want %v", result, test.expected)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicates",
			input:    []string{"192.168.1.1", "10.0.0.1"},
			expected: []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:     "With duplicates",
			input:    []string{"192.168.1.1", "10.0.0.1", "192.168.1.1"},
			expected: []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := removeDuplicates(test.input)
			if len(result) != len(test.expected) {
				t.Errorf("removeDuplicates() length = %v; want %v", len(result), len(test.expected))
			}
			
			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("removeDuplicates()[%d] = %v; want %v", i, v, test.expected[i])
				}
			}
		})
	}
}

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
			handler := http.HandlerFunc(ipv4Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, test.expectedCode)
			}
		})
	}
}

func BenchmarkExtractClientIP(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractClientIP(req)
	}
}

func BenchmarkIsValidIP(b *testing.B) {
	ips := []string{
		"192.168.1.1",
		"2001:db8::1",
		"invalid",
		"127.0.0.1",
		"8.8.8.8",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range ips {
			isValidIP(ip)
		}
	}
}

// Additional comprehensive tests for better coverage

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
			handler := http.HandlerFunc(ipv6Handler)
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
	handler := http.HandlerFunc(infoHandler)
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

func TestHeadersHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/headers", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Custom-Header", "test-value")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(headersHandler)
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
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "No valid IP found",
			headers:      map[string]string{"X-Forwarded-For": "invalid"},
			remoteAddr:   "invalid:12345",
			expectedCode: http.StatusBadRequest,
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
			handler := http.HandlerFunc(ipv4Handler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, test.expectedCode)
			}
		})
	}
}

func TestIsPrivateIPComprehensive(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
		desc     string
	}{
		// IPv4 private ranges
		{"10.0.0.0", true, "10.x.x.x start"},
		{"10.255.255.255", true, "10.x.x.x end"},
		{"172.16.0.0", true, "172.16-31.x.x start"},
		{"172.31.255.255", true, "172.16-31.x.x end"},
		{"192.168.0.0", true, "192.168.x.x start"},
		{"192.168.255.255", true, "192.168.x.x end"},
		{"127.0.0.1", true, "localhost"},
		{"127.255.255.255", true, "loopback range end"},
		{"169.254.0.1", true, "link-local"},
		{"169.254.255.254", true, "link-local end"},
		
		// IPv4 public ranges
		{"8.8.8.8", false, "Google DNS"},
		{"1.1.1.1", false, "Cloudflare DNS"},
		{"203.0.113.1", false, "TEST-NET-3"},
		{"11.0.0.1", false, "just outside 10.x"},
		{"172.15.255.255", false, "just before 172.16"},
		{"172.32.0.0", false, "just after 172.31"},
		{"192.167.255.255", false, "just before 192.168"},
		{"192.169.0.0", false, "just after 192.168"},
		
		// IPv6 private ranges
		{"::1", true, "IPv6 loopback"},
		{"fc00::", true, "unique local start"},
		{"fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true, "unique local end"},
		{"fe80::", true, "link-local start"},
		{"febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true, "link-local end"},
		
		// IPv6 public ranges
		{"2001:db8::", false, "documentation range"},
		{"2001:4860:4860::8888", false, "Google IPv6 DNS"},
		{"2606:4700:4700::1111", false, "Cloudflare IPv6 DNS"},
		{"ff00::", false, "multicast"},
		
		// Edge cases
		{"", false, "empty string"},
		{"invalid", false, "invalid IP"},
		{"999.999.999.999", false, "invalid IPv4"},
		{"gggg::1", false, "invalid IPv6"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := isPrivateIP(test.ip)
			if result != test.expected {
				t.Errorf("isPrivateIP(%s) = %v; want %v (%s)", test.ip, result, test.expected, test.desc)
			}
		})
	}
}

func TestRemoveDuplicatesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Single item",
			input:    []string{"192.168.1.1"},
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "All duplicates",
			input:    []string{"192.168.1.1", "192.168.1.1", "192.168.1.1"},
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "Nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Empty strings",
			input:    []string{"", "test", "", "test2", ""},
			expected: []string{"", "test", "test2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := removeDuplicates(test.input)
			
			if test.expected == nil {
				if result != nil {
					t.Errorf("removeDuplicates() = %v; want nil", result)
				}
				return
			}
			
			if len(result) != len(test.expected) {
				t.Errorf("removeDuplicates() length = %v; want %v", len(result), len(test.expected))
				return
			}
			
			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("removeDuplicates()[%d] = %v; want %v", i, v, test.expected[i])
				}
			}
		})
	}
}

func TestExtractClientIPDetectionVia(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		remoteAddr  string
		expectedIP  string
		expectedVia string
	}{
		{
			name: "True-Client-IP priority",
			headers: map[string]string{
				"True-Client-IP": "203.0.113.1",
				"X-Forwarded-For": "203.0.113.2",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "True-Client-IP",
		},
		{
			name: "X-Client-IP",
			headers: map[string]string{
				"X-Client-IP": "203.0.113.1",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "X-Client-IP",
		},
		{
			name: "X-Cluster-Client-IP",
			headers: map[string]string{
				"X-Cluster-Client-IP": "203.0.113.1",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "X-Cluster-Client-IP",
		},
		{
			name: "Multiple IPs in X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "invalid, 203.0.113.1, 192.168.1.1",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "X-Forwarded-For",
		},
		{
			name: "Invalid IP in header",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "192.168.1.1",
			expectedVia: "RemoteAddr",
		},
		{
			name:        "Malformed RemoteAddr",
			headers:     map[string]string{},
			remoteAddr:  "malformed-addr",
			expectedIP:  "malformed-addr",
			expectedVia: "RemoteAddr",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr
			
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			ip, via := extractClientIP(req)
			if ip != test.expectedIP {
				t.Errorf("extractClientIP() IP = %v; want %v", ip, test.expectedIP)
			}
			if via != test.expectedVia {
				t.Errorf("extractClientIP() via = %v; want %v", via, test.expectedVia)
			}
		})
	}
}

// Benchmark additional functions
func BenchmarkIsPrivateIP(b *testing.B) {
	ips := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"2001:db8::1",
		"::1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range ips {
			isPrivateIP(ip)
		}
	}
}

func BenchmarkFindIPv4(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1, 192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findIPv4(req)
	}
}

func TestJSONHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/json", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Test content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Test JSON structure - need to import encoding/json for this
	var response IPInfo
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

func TestGetIPInfo(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("CF-Ray", "123456789-ABC")
	req.RemoteAddr = "192.168.1.1:12345"

	info := getIPInfo(req)

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

func TestHealthHandlerJSONStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Test JSON structure
	var response HealthResponse
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

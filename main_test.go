package main

import (
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

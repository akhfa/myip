package ip

import (
	"net/http/httptest"
	"testing"
)

func TestIsValid(t *testing.T) {
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
		result := IsValid(test.ip)
		if result != test.expected {
			t.Errorf("IsValid(%s) = %v; want %v", test.ip, result, test.expected)
		}
	}
}

func TestIsPrivate(t *testing.T) {
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
		result := IsPrivate(test.ip)
		if result != test.expected {
			t.Errorf("IsPrivate(%s) = %v; want %v", test.ip, result, test.expected)
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

			ip, _ := ExtractClientIP(req)
			if ip != test.expectedIP {
				t.Errorf("ExtractClientIP() = %v; want %v", ip, test.expectedIP)
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

			result := FindIPv4(req)
			if result != test.expected {
				t.Errorf("FindIPv4() = %v; want %v", result, test.expected)
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

			result := FindIPv6(req)
			if result != test.expected {
				t.Errorf("FindIPv6() = %v; want %v", result, test.expected)
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

			result := IsCloudflareRequest(req)
			if result != test.expected {
				t.Errorf("IsCloudflareRequest() = %v; want %v", result, test.expected)
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
			result := RemoveDuplicates(test.input)
			if len(result) != len(test.expected) {
				t.Errorf("RemoveDuplicates() length = %v; want %v", len(result), len(test.expected))
			}
			
			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("RemoveDuplicates()[%d] = %v; want %v", i, v, test.expected[i])
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkExtractClientIP(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("CF-Connecting-IP", "203.0.113.1")
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractClientIP(req)
	}
}

func BenchmarkIsValid(b *testing.B) {
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
			IsValid(ip)
		}
	}
}

func BenchmarkIsPrivate(b *testing.B) {
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
			IsPrivate(ip)
		}
	}
}

func BenchmarkFindIPv4(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1, 192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindIPv4(req)
	}
}
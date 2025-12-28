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
		name       string
		headers    map[string]string
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

// Additional comprehensive tests for missing coverage

func TestIsPrivateComprehensive(t *testing.T) {
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
			result := IsPrivate(test.ip)
			if result != test.expected {
				t.Errorf("IsPrivate(%s) = %v; want %v (%s)", test.ip, result, test.expected, test.desc)
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
			result := RemoveDuplicates(test.input)

			if test.expected == nil {
				if result != nil {
					t.Errorf("RemoveDuplicates() = %v; want nil", result)
				}
				return
			}

			if len(result) != len(test.expected) {
				t.Errorf("RemoveDuplicates() length = %v; want %v", len(result), len(test.expected))
				return
			}

			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("RemoveDuplicates()[%d] = %v; want %v", i, v, test.expected[i])
				}
			}
		})
	}
}

func TestExtractClientIPEdgeCases(t *testing.T) {
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
				"True-Client-IP":  "203.0.113.1",
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
			name: "X-Forwarded header",
			headers: map[string]string{
				"X-Forwarded": "203.0.113.1",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "X-Forwarded",
		},
		{
			name: "Forwarded-For header",
			headers: map[string]string{
				"Forwarded-For": "203.0.113.1",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "Forwarded-For",
		},
		{
			name: "Forwarded header",
			headers: map[string]string{
				"Forwarded": "203.0.113.1",
			},
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "203.0.113.1",
			expectedVia: "Forwarded",
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

			ip, via := ExtractClientIP(req)
			if ip != test.expectedIP {
				t.Errorf("ExtractClientIP() IP = %v; want %v", ip, test.expectedIP)
			}
			if via != test.expectedVia {
				t.Errorf("ExtractClientIP() via = %v; want %v", via, test.expectedVia)
			}
		})
	}
}

func TestFindIPv4EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "No IPv4 found - only IPv6",
			headers:    map[string]string{"X-Forwarded-For": "2001:db8::1"},
			remoteAddr: "[2001:db8::1]:12345",
			expected:   "",
		},
		{
			name:       "No valid IP found",
			headers:    map[string]string{"X-Forwarded-For": "invalid"},
			remoteAddr: "invalid:12345",
			expected:   "",
		},
		{
			name:       "Malformed RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
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

func TestFindIPv6EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "No IPv6 found - only IPv4",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "",
		},
		{
			name:       "No valid IP found",
			headers:    map[string]string{"X-Forwarded-For": "invalid"},
			remoteAddr: "invalid:12345",
			expected:   "",
		},
		{
			name:       "Malformed RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "2001:db8::1",
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

func TestIsCloudflareRequestTrueClientIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("True-Client-IP", "203.0.113.1")

	if !IsCloudflareRequest(req) {
		t.Error("Expected IsCloudflareRequest to return true for True-Client-IP header")
	}
}

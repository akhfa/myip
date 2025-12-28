package ip

import (
	"log"
	"net"
	"net/http"
	"strings"
)

// Header priority order for IP detection
var headerPriority = []string{
	"CF-Connecting-IP",    // Cloudflare
	"True-Client-IP",      // Cloudflare Enterprise
	"X-Real-IP",           // nginx proxy/FastCGI
	"X-Forwarded-For",     // Standard proxy header
	"X-Client-IP",         // Apache mod_proxy_http
	"X-Cluster-Client-IP", // Cluster environments
	"X-Forwarded",         // Less common
	"Forwarded-For",       // Less common
	"Forwarded",           // Less common
}

// Private IP ranges (IPv4)
var privateIPRanges = []*net.IPNet{
	// RFC 1918
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	// RFC 3927
	parseCIDR("169.254.0.0/16"),
	// RFC 5735
	parseCIDR("127.0.0.0/8"),
}

// Private IPv6 ranges
var privateIPv6Ranges = []*net.IPNet{
	// RFC 4193 - Unique Local Addresses
	parseCIDR("fc00::/7"),
	// RFC 4291 - Link-Local
	parseCIDR("fe80::/10"),
	// RFC 4291 - Loopback
	parseCIDR("::1/128"),
}

func parseCIDR(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("Failed to parse CIDR %s: %v", cidr, err)
	}
	return network
}

// IsValid checks if the given string is a valid IP address
func IsValid(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsPrivate checks if the given IP address is in a private range
func IsPrivate(ip string) bool {
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check IPv4 private ranges
	if parsedIP.To4() != nil {
		for _, privateRange := range privateIPRanges {
			if privateRange.Contains(parsedIP) {
				return true
			}
		}
		return false
	}

	// Check IPv6 private ranges
	for _, privateRange := range privateIPv6Ranges {
		if privateRange.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// IsCloudflareRequest checks if the request comes through Cloudflare
func IsCloudflareRequest(r *http.Request) bool {
	return r.Header.Get("CF-Connecting-IP") != "" ||
		r.Header.Get("CF-Ray") != "" ||
		r.Header.Get("True-Client-IP") != ""
}

// ExtractClientIP extracts the client IP from request headers with detection method
func ExtractClientIP(r *http.Request) (string, string) {
	// Check headers in priority order
	for _, header := range headerPriority {
		value := r.Header.Get(header)
		if value != "" {
			// Handle comma-separated IPs (take the first valid one)
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if IsValid(ip) {
					return ip, header
				}
			}
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, "RemoteAddr"
	}
	return host, "RemoteAddr"
}

// FindIPv4 finds the first valid IPv4 address from the request
func FindIPv4(r *http.Request) string {
	// Check headers in priority order
	for _, header := range headerPriority {
		value := r.Header.Get(header)
		if value != "" {
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if IsValid(ip) {
					parsedIP := net.ParseIP(ip)
					if parsedIP != nil && parsedIP.To4() != nil {
						return ip
					}
				}
			}
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if IsValid(host) {
		parsedIP := net.ParseIP(host)
		if parsedIP != nil && parsedIP.To4() != nil {
			return host
		}
	}

	return ""
}

// FindIPv6 finds the first valid IPv6 address from the request
func FindIPv6(r *http.Request) string {
	// Check headers in priority order
	for _, header := range headerPriority {
		value := r.Header.Get(header)
		if value != "" {
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if IsValid(ip) {
					parsedIP := net.ParseIP(ip)
					if parsedIP != nil && parsedIP.To4() == nil {
						return ip
					}
				}
			}
		}
	}

	// Fall back to RemoteAddr (handle bracketed IPv6)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if IsValid(host) {
		parsedIP := net.ParseIP(host)
		if parsedIP != nil && parsedIP.To4() == nil {
			return host
		}
	}

	return ""
}

// RemoveDuplicates removes duplicate strings from a slice while preserving order
func RemoveDuplicates(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

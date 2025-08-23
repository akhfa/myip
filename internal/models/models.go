package models

import "time"

// IPInfo represents detailed information about the client's IP
type IPInfo struct {
	ClientIP     string `json:"client_ip"`
	DetectedVia  string `json:"detected_via"`
	IPv4Address  string `json:"ipv4_address"`
	IPv6Address  string `json:"ipv6_address"`
	IsPrivateIP  bool   `json:"is_private_ip"`
	IsCloudflare bool   `json:"is_cloudflare"`
	UserAgent    string `json:"user_agent"`
	Timestamp    string `json:"timestamp"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// NewHealthResponse creates a new health response with current timestamp
func NewHealthResponse(status string) *HealthResponse {
	return &HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

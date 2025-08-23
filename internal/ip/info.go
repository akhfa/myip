package ip

import (
	"net/http"
	"time"

	"myip/internal/models"
)

// GetInfo gets comprehensive IP information
func GetInfo(r *http.Request) *models.IPInfo {
	clientIP, detectedVia := ExtractClientIP(r)
	ipv4 := FindIPv4(r)
	ipv6 := FindIPv6(r)

	return &models.IPInfo{
		ClientIP:     clientIP,
		DetectedVia:  detectedVia,
		IPv4Address:  ipv4,
		IPv6Address:  ipv6,
		IsPrivateIP:  IsPrivate(clientIP),
		IsCloudflare: IsCloudflareRequest(r),
		UserAgent:    r.Header.Get("User-Agent"),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
}

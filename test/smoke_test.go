package smoke_test

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"myip/internal/models"
)

const (
	// Target deployment URL for smoke tests
	smokeTestURL = "https://ip.2ak.me"
	// External IP detection services
	ipifyIPv4URL = "https://api.ipify.org"
	ipifyIPv6URL = "https://api64.ipify.org"
	// HTTP client timeout for smoke tests
	smokeTestTimeout = 15 * time.Second
)

// getPublicIP gets the public IP from ipify.org
func getPublicIP(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

// getMyIPResponse gets the IP response from our deployment
func getMyIPResponse(client *http.Client, endpoint string) (string, error) {
	url := smokeTestURL + endpoint
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

// TestSmokeTest is the main smoke test that validates IP detection accuracy
// Run with: go test -run TestSmokeTest -v ./test
func TestSmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	t.Log("=== MANUAL SMOKE TEST TRIGGER ===")
	t.Logf("Testing deployed application at: %s", smokeTestURL)
	t.Log("Validating IP detection accuracy against external IP services...")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: smokeTestTimeout,
	}

	// Test IPv4 detection
	t.Run("IPv4Detection", func(t *testing.T) {
		t.Log("Testing IPv4 detection accuracy...")

		// Get actual IPv4 from ipify
		actualIPv4, err := getPublicIP(client, ipifyIPv4URL)
		if err != nil {
			t.Fatalf("❌ Failed to get IPv4 from ipify.org: %v", err)
		}
		t.Logf("✅ Retrieved IPv4 from ipify.org: %s", actualIPv4)

		// Get IPv4 from our deployment
		detectedIPv4, err := getMyIPResponse(client, "/")
		if err != nil {
			t.Fatalf("❌ Failed to get IPv4 from deployment: %v", err)
		}
		t.Logf("✅ Retrieved IPv4 from deployment: %s", detectedIPv4)

		// Compare results - must match exactly
		if actualIPv4 != detectedIPv4 {
			t.Errorf("❌ IPv4 DETECTION FAILED")
			t.Errorf("   Expected (ipify.org): %s", actualIPv4)
			t.Errorf("   Actual (deployment):  %s", detectedIPv4)
			t.Errorf("   ❗ IPs must match exactly - deployment is not detecting correct client IP")
		} else {
			t.Logf("✅ IPv4 detection SUCCESS: %s matches expected", detectedIPv4)
		}
	})

	// Test IPv6 detection
	t.Run("IPv6Detection", func(t *testing.T) {
		t.Log("Testing IPv6 detection accuracy...")

		// Get actual IPv6 from ipify
		actualIPv6, err := getPublicIP(client, ipifyIPv6URL)
		if err != nil {
			t.Logf("⚠️  Failed to get IPv6 from ipify.org (may not have IPv6): %v", err)
			t.Log("Skipping IPv6 test - no IPv6 connectivity available")
			return
		}
		t.Logf("Actual IPv6 from ipify.org: %s", actualIPv6)

		// Check if ipify returned an IPv4 address (indicating no IPv6 connectivity)
		ip := net.ParseIP(actualIPv6)
		if ip == nil || ip.To4() != nil {
			t.Logf("⚠️  ipify.org returned IPv4 (%s) instead of IPv6 - no IPv6 connectivity available", actualIPv6)

			// Our deployment should also return 404 or "No IPv6 address found"
			resp, err := client.Get(smokeTestURL + "/ipv6")
			if err != nil {
				t.Fatalf("Failed to access /ipv6 endpoint: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				t.Log("✅ IPv6 detection SUCCESS: Deployment correctly returns 404 when no IPv6 available")
			} else {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read IPv6 endpoint response: %v", err)
				}
				bodyStr := strings.TrimSpace(string(body))
				if strings.Contains(bodyStr, "No IPv6") || strings.Contains(bodyStr, "not found") {
					t.Logf("✅ IPv6 detection SUCCESS: Deployment correctly indicates no IPv6: %s", bodyStr)
				} else {
					t.Errorf("❌ IPv6 ENDPOINT BEHAVIOR FAILED")
					t.Errorf("   Expected: 404 status or 'No IPv6' message")
					t.Errorf("   Actual: %s", bodyStr)
					t.Errorf("   ❗ Deployment should indicate when IPv6 is not available")
				}
			}
			return
		}

		// We have real IPv6, test normally
		detectedIPv6, err := getMyIPResponse(client, "/ipv6")
		if err != nil {
			t.Fatalf("Failed to get IPv6 from deployment: %v", err)
		}
		t.Logf("Detected IPv6 from deployment: %s", detectedIPv6)

		// Compare results - must match exactly
		if actualIPv6 != detectedIPv6 {
			t.Errorf("❌ IPv6 DETECTION FAILED")
			t.Errorf("   Expected (ipify.org): %s", actualIPv6)
			t.Errorf("   Actual (deployment):  %s", detectedIPv6)
			t.Errorf("   ❗ IPs must match exactly - deployment is not detecting correct client IP")
		} else {
			t.Logf("✅ IPv6 detection SUCCESS: %s matches expected", detectedIPv6)
		}
	})

	// Test JSON endpoint with detailed comparison
	t.Run("JSONEndpointValidation", func(t *testing.T) {
		t.Log("Testing JSON endpoint with IP comparison...")

		// Re-retrieve actual IPv4 from ipify before JSON comparison
		actualIPv4, err := getPublicIP(client, ipifyIPv4URL)
		if err != nil {
			t.Fatalf("❌ Failed to re-retrieve IPv4 from ipify.org: %v", err)
		}
		t.Logf("✅ Fresh IPv4 from ipify.org for JSON test: %s", actualIPv4)

		// Also get IPv6 for validation
		actualIPv6, err := getPublicIP(client, ipifyIPv6URL)
		if err != nil {
			t.Logf("ℹ️  Could not get IPv6 from ipify.org: %v", err)
			actualIPv6 = ""
		} else {
			// Use proper IP parsing
			ip := net.ParseIP(actualIPv6)
			if ip == nil || ip.To4() != nil {
				t.Logf("ℹ️  No IPv6 connectivity (ipify returned IPv4: %s)", actualIPv6)
				actualIPv6 = ""
			} else {
				t.Logf("✅ Fresh IPv6 from ipify.org for JSON test: %s", actualIPv6)
			}
		}

		// Get JSON response from deployment
		resp, err := client.Get(smokeTestURL + "/json")
		if err != nil {
			t.Fatalf("Failed to access /json endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("❌ JSON endpoint returned status %d, expected 200", resp.StatusCode)
		}
		t.Logf("✅ JSON endpoint accessible (status: %d)", resp.StatusCode)

		// Check content type
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("❌ JSON CONTENT TYPE FAILED")
			t.Errorf("   Expected: application/json")
			t.Errorf("   Actual: %s", ct)
		} else {
			t.Logf("✅ JSON content type correct: %s", ct)
		}

		// Parse JSON response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read JSON response: %v", err)
		}

		var jsonResponse models.IPInfo
		if err := json.Unmarshal(body, &jsonResponse); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		t.Logf("JSON Response - Client IP: %s, Detected Via: %s", jsonResponse.ClientIP, jsonResponse.DetectedVia)

		// Compare IPv4 address field - must match exactly
		if jsonResponse.IPv4Address != actualIPv4 {
			t.Errorf("❌ JSON IPv4 FIELD DETECTION FAILED")
			t.Errorf("   Expected (ipify.org): %s", actualIPv4)
			t.Errorf("   Actual (ipv4_address field): %s", jsonResponse.IPv4Address)
			t.Errorf("   Detection method: %s", jsonResponse.DetectedVia)
			t.Errorf("   ❗ JSON ipv4_address field must match ipify.org result")
		} else {
			t.Logf("✅ JSON IPv4 field detection SUCCESS: %s matches expected", jsonResponse.IPv4Address)
		}

		// Validate IPv6 address field if we have IPv6 connectivity
		if actualIPv6 != "" {
			if jsonResponse.IPv6Address != actualIPv6 {
				t.Errorf("❌ JSON IPv6 FIELD DETECTION FAILED")
				t.Errorf("   Expected (ipify.org): %s", actualIPv6)
				t.Errorf("   Actual (ipv6_address field): %s", jsonResponse.IPv6Address)
				t.Errorf("   ❗ JSON ipv6_address field must match ipify.org result")
			} else {
				t.Logf("✅ JSON IPv6 field detection SUCCESS: %s matches expected", jsonResponse.IPv6Address)
			}
		} else {
			// No IPv6 connectivity, just log the IPv6 field value
			t.Logf("ℹ️  IPv6 field (no IPv6 connectivity to validate): %s", jsonResponse.IPv6Address)
		}

		// Log client_ip for reference but don't validate it
		t.Logf("ℹ️  Client IP field (for reference): %s", jsonResponse.ClientIP)

		// Validate JSON structure and required fields
		if jsonResponse.DetectedVia == "" {
			t.Errorf("❌ JSON STRUCTURE VALIDATION FAILED")
			t.Errorf("   DetectedVia field is empty - JSON response is missing required field")
		} else {
			t.Logf("✅ DetectedVia field populated: %s", jsonResponse.DetectedVia)
		}

		if jsonResponse.Timestamp == "" {
			t.Errorf("❌ JSON STRUCTURE VALIDATION FAILED")
			t.Errorf("   Timestamp field is empty - JSON response is missing required field")
		} else {
			t.Logf("✅ Timestamp field populated: %s", jsonResponse.Timestamp)
		}

		t.Logf("✅ JSON endpoint validation completed successfully")
	})

	// Test basic endpoint accessibility
	t.Run("EndpointAccessibility", func(t *testing.T) {
		t.Log("Testing basic endpoint accessibility...")

		endpoints := []string{"/health", "/info", "/headers"}
		for _, endpoint := range endpoints {
			resp, err := client.Get(smokeTestURL + endpoint)
			if err != nil {
				t.Errorf("❌ ENDPOINT CONNECTION FAILED")
				t.Errorf("   Endpoint: %s", endpoint)
				t.Errorf("   Error: %v", err)
				t.Errorf("   ❗ Cannot connect to endpoint - network or server issue")
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("❌ ENDPOINT ACCESSIBILITY FAILED")
				t.Errorf("   Endpoint: %s", endpoint)
				t.Errorf("   Expected status: 200")
				t.Errorf("   Actual status: %d", resp.StatusCode)
				t.Errorf("   ❗ Endpoint is not accessible or returning errors")
			} else {
				t.Logf("✅ Endpoint %s accessible (status: %d)", endpoint, resp.StatusCode)
			}
		}
	})

	t.Log("=== SMOKE TEST COMPLETED ===")
	t.Log("Deployment validation results:")
	t.Log("  ✅ IPv4 detection accuracy verified")
	t.Log("  ✅ IPv6 detection tested (if available)")
	t.Log("  ✅ JSON endpoint validation completed")
	t.Log("  ✅ Basic endpoint accessibility confirmed")
	t.Log("=== LIVE DEPLOYMENT VALIDATED ===")
}
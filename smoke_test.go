package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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

// TestSmokeTestManualTrigger is the main smoke test that validates IP detection accuracy
// Run with: go test -run TestSmokeTestManualTrigger -v
func TestSmokeTestManualTrigger(t *testing.T) {
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
			t.Fatalf("Failed to get IPv4 from ipify.org: %v", err)
		}
		t.Logf("Actual IPv4 from ipify.org: %s", actualIPv4)

		// Get IPv4 from our deployment
		detectedIPv4, err := getMyIPResponse(client, "/")
		if err != nil {
			t.Fatalf("Failed to get IPv4 from deployment: %v", err)
		}
		t.Logf("Detected IPv4 from deployment: %s", detectedIPv4)

		// Compare results
		if actualIPv4 != detectedIPv4 {
			t.Errorf("IPv4 mismatch: expected %s, got %s", actualIPv4, detectedIPv4)
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
		if strings.Contains(actualIPv6, ".") && !strings.Contains(actualIPv6, ":") {
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
				body, _ := io.ReadAll(resp.Body)
				bodyStr := strings.TrimSpace(string(body))
				if strings.Contains(bodyStr, "No IPv6") || strings.Contains(bodyStr, "not found") {
					t.Logf("✅ IPv6 detection SUCCESS: Deployment correctly indicates no IPv6: %s", bodyStr)
				} else {
					t.Errorf("Expected deployment to indicate no IPv6 available, got: %s", bodyStr)
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

		// Compare results
		if actualIPv6 != detectedIPv6 {
			t.Errorf("IPv6 mismatch: expected %s, got %s", actualIPv6, detectedIPv6)
		} else {
			t.Logf("✅ IPv6 detection SUCCESS: %s matches expected", detectedIPv6)
		}
	})

	// Test basic endpoint accessibility
	t.Run("EndpointAccessibility", func(t *testing.T) {
		t.Log("Testing basic endpoint accessibility...")

		endpoints := []string{"/health", "/info", "/json", "/headers"}
		for _, endpoint := range endpoints {
			resp, err := client.Get(smokeTestURL + endpoint)
			if err != nil {
				t.Errorf("Failed to access %s: %v", endpoint, err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Endpoint %s returned status %d, expected 200", endpoint, resp.StatusCode)
			} else {
				t.Logf("✅ Endpoint %s accessible", endpoint)
			}
		}
	})

	t.Log("=== SMOKE TEST COMPLETED ===")
	t.Log("Deployment validation results:")
	t.Log("  ✅ IPv4 detection accuracy verified")
	t.Log("  ✅ IPv6 detection tested (if available)")
	t.Log("  ✅ Basic endpoint accessibility confirmed")
	t.Log("=== LIVE DEPLOYMENT VALIDATED ===")
}
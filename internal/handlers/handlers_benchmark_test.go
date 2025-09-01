package handlers

import (
	"strings"
	"testing"
)

// BenchmarkIsJSONFormat benchmarks our optimized case-insensitive comparison
func BenchmarkIsJSONFormat(b *testing.B) {
	testCases := []string{
		"json",
		"JSON",
		"Json",
		"jSoN",
		"xml",
		"text",
		"",
		"jsonformat", // longer string
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, format := range testCases {
			_ = isJSONFormat(format)
		}
	}
}

// BenchmarkStringToLower benchmarks the strings.ToLower approach
func BenchmarkStringToLower(b *testing.B) {
	testCases := []string{
		"json",
		"JSON",
		"Json",
		"jSoN",
		"xml",
		"text",
		"",
		"jsonformat", // longer string
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, format := range testCases {
			_ = strings.ToLower(format) == "json"
		}
	}
}

// BenchmarkHandlerJSONCheck benchmarks the format check in context
func BenchmarkHandlerJSONCheck(b *testing.B) {
	formats := []string{"json", "JSON", "Json", "xml", ""}

	b.Run("Optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, format := range formats {
				_ = isJSONFormat(format)
			}
		}
	})

	b.Run("StringsToLower", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, format := range formats {
				_ = strings.ToLower(format) == "json"
			}
		}
	})
}

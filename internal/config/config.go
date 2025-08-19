package config

import "os"

// Config holds application configuration
type Config struct {
	Port string
}

// Load loads configuration from environment variables
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	return &Config{
		Port: port,
	}
}

// GetAddr returns the server address string
func (c *Config) GetAddr() string {
	return ":" + c.Port
}
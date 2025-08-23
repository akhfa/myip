package config

import "os"

// Config holds application configuration
type Config struct {
	Port string
	Host string
}

// Load loads configuration from environment variables
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost:8080"
	}
	
	return &Config{
		Port: port,
		Host: host,
	}
}

// GetAddr returns the server address string
func (c *Config) GetAddr() string {
	return ":" + c.Port
}
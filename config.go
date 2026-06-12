package main

import "os"

type Config struct {
	APIKey string // PROXYSEO_API_KEY — service API key, identifies the service
	APIURL string // PROXYSEO_API_URL — API base URL
}

func LoadConfig() Config {
	apiURL := os.Getenv("PROXYSEO_API_URL")
	if apiURL == "" {
		apiURL = "https://api.proxyseo.es"
	}
	return Config{
		APIKey: os.Getenv("PROXYSEO_API_KEY"),
		APIURL: apiURL,
	}
}

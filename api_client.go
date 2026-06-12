package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient talks to the ProxySEO API v2 client endpoints using a service
// API key. The key identifies the service server-side, so no service ID is
// needed. Only two read-only endpoints are consumed:
//
//	GET /api/v2/client/service           — service info (ports, creds, plan)
//	GET /api/v2/client/service/addresses — assigned IPs + ports + creds
type APIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewAPIClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *APIClient) doRequest(method, path string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized: invalid or revoked API key (check PROXYSEO_API_KEY)")
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limit exceeded (60 req/min), retry later")
	case resp.StatusCode >= 400:
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// ServiceInfo mirrors GET /api/v2/client/service.
type ServiceInfo struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Status      string `json:"status"`
	AuthType    string `json:"auth_type"`
	PortHTTP    int    `json:"port_http"`
	PortSOCKS   int    `json:"port_socks"`
	ProductName string `json:"product_name"`
	IPCount     int    `json:"ip_count"`
	ProxyUser   string `json:"proxy_user"`
	ProxyPass   string `json:"proxy_pass"`
}

// AddressEntry mirrors GET /api/v2/client/service/addresses items.
type AddressEntry struct {
	ID        int64  `json:"id"`
	IP        string `json:"ip"`
	PortHTTP  int    `json:"port_http"`
	PortSOCKS int    `json:"port_socks"`
	ProxyUser string `json:"proxy_user"`
	ProxyPass string `json:"proxy_pass"`
}

// GetService returns the service that owns the API key.
func (c *APIClient) GetService() (*ServiceInfo, error) {
	data, err := c.doRequest("GET", "/api/v2/client/service")
	if err != nil {
		return nil, err
	}
	var svc ServiceInfo
	if err := json.Unmarshal(data, &svc); err != nil {
		return nil, fmt.Errorf("decode service: %w", err)
	}
	return &svc, nil
}

// GetAddresses returns the assigned IPs (with ports and credentials).
func (c *APIClient) GetAddresses() ([]AddressEntry, error) {
	data, err := c.doRequest("GET", "/api/v2/client/service/addresses")
	if err != nil {
		return nil, err
	}
	var addrs []AddressEntry
	if err := json.Unmarshal(data, &addrs); err != nil {
		return nil, fmt.Errorf("decode addresses: %w", err)
	}
	return addrs, nil
}

package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

const version = "1.1.0"

func main() {
	cfg := LoadConfig()

	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "ERROR: PROXYSEO_API_KEY environment variable is required")
		fmt.Fprintln(os.Stderr, "Get your API key from your ProxySEO service panel or support.")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "ProxySEO MCP Server v%s starting...\n", version)
	fmt.Fprintf(os.Stderr, "API: %s\n", cfg.APIURL)

	client := NewAPIClient(cfg.APIURL, cfg.APIKey)

	s := server.NewMCPServer(
		"proxyseo-mcp",
		version,
		server.WithToolCapabilities(false),
	)

	registerTools(s, client)

	fmt.Fprintln(os.Stderr, "Tools registered: get_proxy, list_proxies, proxy_status, check_proxy")
	fmt.Fprintln(os.Stderr, "Ready. Listening on stdio...")

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

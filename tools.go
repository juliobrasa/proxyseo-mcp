package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MVP tool set. Every tool is backed exclusively by the two client
// endpoints (/api/v2/client/service[+/addresses]). Tools that required
// admin-only endpoints (rotate_proxy, update_source_ip) are NOT exposed.
func registerTools(s mcpServer, client *APIClient) {
	// ── get_proxy ──
	getProxyTool := mcp.NewTool("get_proxy",
		mcp.WithDescription("Get a proxy ready to use (IP:port + credentials). Returns connection URL or JSON."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("protocol",
			mcp.Description("Protocol: http or socks5"),
			mcp.Enum("http", "socks5"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: url or json"),
			mcp.Enum("url", "json"),
		),
	)
	s.AddTool(getProxyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		protocol := req.GetString("protocol", "http")
		format := req.GetString("format", "url")

		addrs, err := client.GetAddresses()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get addresses: %v", err)), nil
		}
		if len(addrs) == 0 {
			return mcp.NewToolResultError("No proxy IPs assigned to this service"), nil
		}
		a := addrs[0]

		if format == "json" {
			result := map[string]interface{}{
				"ip":         a.IP,
				"http_port":  a.PortHTTP,
				"socks_port": a.PortSOCKS,
				"username":   a.ProxyUser,
				"password":   a.ProxyPass,
				"protocol":   protocol,
			}
			b, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(b)), nil
		}

		port := a.PortHTTP
		scheme := "http"
		if protocol == "socks5" {
			port = a.PortSOCKS
			scheme = "socks5"
		}
		proxyURL := fmt.Sprintf("%s://%s:%s@%s:%d", scheme, url.PathEscape(a.ProxyUser), url.PathEscape(a.ProxyPass), a.IP, port)
		return mcp.NewToolResultText(proxyURL), nil
	})

	// ── list_proxies ──
	listProxiesTool := mcp.NewTool("list_proxies",
		mcp.WithDescription("List all proxy IPs assigned to your service."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("format",
			mcp.Description("Output format: simple (IP:port per line) or detailed (JSON with ports and credentials)"),
			mcp.Enum("simple", "detailed"),
		),
	)
	s.AddTool(listProxiesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format := req.GetString("format", "simple")

		addrs, err := client.GetAddresses()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list addresses: %v", err)), nil
		}
		if len(addrs) == 0 {
			return mcp.NewToolResultText("No proxy IPs assigned"), nil
		}

		if format == "detailed" {
			b, _ := json.MarshalIndent(addrs, "", "  ")
			return mcp.NewToolResultText(string(b)), nil
		}

		var sb strings.Builder
		for _, a := range addrs {
			fmt.Fprintf(&sb, "%s:%d\n", a.IP, a.PortHTTP)
		}
		return mcp.NewToolResultText(sb.String()), nil
	})

	// ── proxy_status ──
	statusTool := mcp.NewTool("proxy_status",
		mcp.WithDescription("Get the status of your proxy service (plan, auth type, IP count, ports)."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
	)
	s.AddTool(statusTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := client.GetService()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get service: %v", err)), nil
		}

		addrs, err := client.GetAddresses()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list addresses: %v", err)), nil
		}

		result := map[string]interface{}{
			"service_id":  svc.ID,
			"status":      svc.Status,
			"product":     svc.ProductName,
			"plan_ips":    svc.IPCount,
			"active_ips":  len(addrs),
			"auth_method": svc.AuthType,
			"http_port":   svc.PortHTTP,
			"socks_port":  svc.PortSOCKS,
		}
		b, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	// ── check_proxy ──
	// Tests connectivity THROUGH the proxy directly (no admin endpoint needed).
	checkTool := mcp.NewTool("check_proxy",
		mcp.WithDescription("Verify that a proxy is working (connectivity + anonymity check through the proxy)."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithString("ip",
			mcp.Description("Specific proxy IP to check. If empty, checks the first assigned proxy."),
		),
	)
	s.AddTool(checkTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		targetIP := req.GetString("ip", "")

		addrs, err := client.GetAddresses()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get addresses: %v", err)), nil
		}
		if len(addrs) == 0 {
			return mcp.NewToolResultError("No proxy IPs assigned"), nil
		}

		target := addrs[0]
		if targetIP != "" {
			found := false
			for _, a := range addrs {
				if a.IP == targetIP {
					target = a
					found = true
					break
				}
			}
			if !found {
				return mcp.NewToolResultError(fmt.Sprintf("IP %s does not belong to your service", targetIP)), nil
			}
		}

		proxyURLStr := fmt.Sprintf("http://%s:%s@%s:%d",
			url.PathEscape(target.ProxyUser), url.PathEscape(target.ProxyPass), target.IP, target.PortHTTP)
		proxyURL, _ := url.Parse(proxyURLStr)

		transport := &http.Transport{
			Proxy:             http.ProxyURL(proxyURL),
			DisableKeepAlives: true,
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
		}
		httpClient := &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		}

		start := time.Now()
		resp, err := httpClient.Get("https://api.ipify.org?format=json")
		elapsed := time.Since(start)

		if err != nil {
			result := map[string]interface{}{
				"ip":     target.IP,
				"status": "error",
				"error":  err.Error(),
			}
			b, _ := json.MarshalIndent(result, "", "  ")
			return mcp.NewToolResultText(string(b)), nil
		}
		defer resp.Body.Close()

		var body struct {
			IP string `json:"ip"`
		}
		json.NewDecoder(resp.Body).Decode(&body)

		result := map[string]interface{}{
			"ip":               target.IP,
			"status":           "ok",
			"response_time_ms": elapsed.Milliseconds(),
			"external_ip":      body.IP,
			"anonymous":        body.IP == target.IP,
		}
		b, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})
}

// mcpServer interface for AddTool
type mcpServer interface {
	AddTool(tool mcp.Tool, handler server.ToolHandlerFunc)
}

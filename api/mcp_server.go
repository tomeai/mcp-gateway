package api

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tomeai/mcp-gateway/model"
	"github.com/tomeai/mcp-gateway/service"
	"net/http"
	"time"
)

type DynamicMCPServer struct {
	mcpSProxyServer *server.MCPServer
}

func NewDynamicMCPServer() *DynamicMCPServer {
	// load from db by uid && mcpServerName
	return &DynamicMCPServer{}
}

func (m *DynamicMCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mcpTransport := r.PathValue("type")
	if mcpTransport != "mcp" && mcpTransport != "http" {
		http.Error(w, "invalid mcpTransport", http.StatusBadRequest)
		return
	}

	mcpServerName := r.PathValue("name")
	if mcpServerName == "" {
		http.Error(w, "mcpServerName is nil", http.StatusBadRequest)
		return
	}

	timeCtx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	clientConfig := &model.MCPClientConfig{}
	err := json.Unmarshal([]byte("{\"command\":\"npx\",\"args\":[\"-y\",\"@modelcontextprotocol/server-github\"],\"env\":{\"GITHUB_PERSONAL_ACCESS_TOKEN\":\"<YOUR_TOKEN>\"},\"options\":{\"toolFilter\":{\"mode\":\"block\",\"list\":[\"create_or_update_file\"]}}}"), clientConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mcpClient, err := service.NewMCPClientService("wemcp-gateway", clientConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// server: streamable http && sse
	mcpProxyServer := server.NewMCPServer(
		mcpServerName,
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	//
	err = mcpClient.AddToMCPServer(timeCtx, mcpProxyServer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set up the MCP proxy server on /mcp
	if mcpTransport == "http" {
		server.NewStreamableHTTPServer(
			mcpProxyServer,
			server.WithStateLess(true),
		).ServeHTTP(w, r)

	} else {
		// Set up the SSE transport-based MCP proxy server for the global /sse endpoint
		server.NewSSEServer(
			mcpProxyServer,
			server.WithStaticBasePath(mcpServerName),
		).ServeHTTP(w, r)
	}
}

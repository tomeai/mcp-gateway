package service

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tomeai/mcp-gateway/model"
	"github.com/tomeai/mcp-gateway/repository"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type DynamicMCPServer struct {
	mcpServerService *repository.McpServerService
	logger           *zap.Logger
}

func NewDynamicMCPServer(mcpServerService *repository.McpServerService, logger *zap.Logger) *DynamicMCPServer {
	// load from db by uid && mcpServerName
	return &DynamicMCPServer{
		mcpServerService: mcpServerService,
		logger:           logger,
	}
}

func (m *DynamicMCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mcpServerName := r.PathValue("name")
	if mcpServerName == "" {
		http.Error(w, "mcpServerName is nil", http.StatusBadRequest)
		return
	}

	m.logger.Info("dynamic mcp", zap.String("mcpServerName", mcpServerName))
	mcpServerConfig, err := m.mcpServerService.GetMcpServer("gage", mcpServerName)
	// store

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeCtx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	clientConfig := &model.MCPClientConfig{}
	err = sonic.Unmarshal(mcpServerConfig.ServerConfig, clientConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mcpClient, err := NewMCPClientService("wemcp-gateway", clientConfig, m.logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// server: streamable http
	mcpProxyServer := server.NewMCPServer(
		mcpServerName,
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	// add mcp server
	err = mcpClient.AddToMCPServer(timeCtx, mcpProxyServer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	server.NewStreamableHTTPServer(
		mcpProxyServer,
		server.WithStateLess(true),
	).ServeHTTP(w, r)
}

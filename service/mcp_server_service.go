package service

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tomeai/mcp-gateway/model"
	"github.com/tomeai/mcp-gateway/repository"
	"github.com/tomeai/mcp-gateway/utils"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type DynamicMCPServer struct {
	mcpServerService *repository.McpServerService
	mcpServerMcp     sync.Map
	logger           *zap.Logger
}

func NewDynamicMCPServer(mcpServerService *repository.McpServerService, logger *zap.Logger) *DynamicMCPServer {
	// load from db by uid && mcpServerName
	return &DynamicMCPServer{
		mcpServerService: mcpServerService,
		logger:           logger,
	}
}

func (m *DynamicMCPServer) buildMcpServer(mcpServer *model.McpServer) (*server.MCPServer, error) {
	timeCtx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	clientConfig := &model.MCPClientConfig{}
	err := sonic.Unmarshal(mcpServer.ServerConfig, clientConfig)
	if err != nil {
		return nil, err
	}
	mcpClient, err := NewMCPClientService("wemcp-gateway", clientConfig, m.logger)
	if err != nil {
		return nil, err
	}
	// server: streamable http
	mcpProxyServer := server.NewMCPServer(
		mcpServer.ServerName,
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	// add mcp server
	err = mcpClient.AddToMCPServer(timeCtx, mcpProxyServer)
	if err != nil {
		return nil, err
	}
	return mcpProxyServer, nil
}

func (m *DynamicMCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		mcpProxyServer *server.MCPServer
		err            error
	)

	mcpServerName := r.PathValue("name")
	if mcpServerName == "" {
		http.Error(w, "mcpServerName is nil", http.StatusBadRequest)
		return
	}

	m.logger.Info("dynamic mcp", zap.String("mcpServerName", mcpServerName))
	mcpServer, err := m.mcpServerService.GetMcpServer("gage", mcpServerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	serverMd5 := utils.Md5String(string(mcpServer.ServerConfig))
	m.logger.Info("serverMd5", zap.String("serverMd5", serverMd5), zap.String("serverConfig", string(mcpServer.ServerConfig)))
	if v, ok := m.mcpServerMcp.Load(serverMd5); !ok {
		// 构建
		mcpProxyServer, err = m.buildMcpServer(mcpServer)
		m.mcpServerMcp.Store(serverMd5, mcpProxyServer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		mcpProxyServer = v.(*server.MCPServer)
	}

	server.NewStreamableHTTPServer(
		mcpProxyServer,
		server.WithStateLess(true),
	).ServeHTTP(w, r)
}

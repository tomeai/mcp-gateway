package api

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tomeai/mcp-gateway/model"
	"github.com/tomeai/mcp-gateway/service"
	"github.com/urfave/cli/v2"
	"log"
	"time"
)

type MCPAPIServer struct {
	sseHandler  *server.SSEServer
	httpHandler *server.StreamableHTTPServer
}

func NewMCPServer(ctx *cli.Context, uid, mcpServerName string) *MCPAPIServer {
	// load from db by uid && mcpServerName
	// {"github":{"command":"npx","args":["-y","@modelcontextprotocol/server-github"],"env":{"GITHUB_PERSONAL_ACCESS_TOKEN":"<YOUR_TOKEN>"},"options":{"toolFilter":{"mode":"block","list":["create_or_update_file"]}}}}

	timeCtx, cancel := context.WithTimeout(ctx.Context, time.Minute*3)
	defer cancel()
	name := "github"
	clientConfig := &model.MCPClientConfig{}
	err := json.Unmarshal([]byte("{\"command\":\"npx\",\"args\":[\"-y\",\"@modelcontextprotocol/server-github\"],\"env\":{\"GITHUB_PERSONAL_ACCESS_TOKEN\":\"<YOUR_TOKEN>\"},\"options\":{\"toolFilter\":{\"mode\":\"block\",\"list\":[\"create_or_update_file\"]}}}"), clientConfig)
	if err != nil {
		log.Fatal(err)
	}
	mcpClient, err := service.NewMCPClientService(name, clientConfig)
	if err != nil {
		log.Fatal(err)
	}
	// server: streamable http && sse
	mcpProxyServer := server.NewMCPServer(
		name,
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	//
	err = mcpClient.AddToMCPServer(timeCtx, mcpProxyServer)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the MCP proxy server on /mcp
	streamableHTTPServer := server.NewStreamableHTTPServer(
		mcpProxyServer,
		server.WithStateLess(true),
	)

	// Set up the SSE transport-based MCP proxy server for the global /sse endpoint
	sseServer := server.NewSSEServer(
		mcpProxyServer,
		server.WithStaticBasePath(name),
		//server.WithBaseURL("http://127.0.0.1:8000"),
	)

	return &MCPAPIServer{
		sseHandler:  sseServer,
		httpHandler: streamableHTTPServer,
	}
}

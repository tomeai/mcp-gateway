package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tomeai/mcp-gateway/internal/telemetry"
	"github.com/tomeai/mcp-gateway/repository"
	"github.com/tomeai/mcp-gateway/service"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	ctx *cli.Context
	*http.Server

	mcpClientService *repository.McpClientService

	dynamicMCPServer *service.DynamicMCPServer

	otelProviders *telemetry.Providers
	metrics       telemetry.CustomMetrics

	logger *zap.Logger
}

func NewOtel(ctx *cli.Context) (*telemetry.Providers, error) {
	otelConfig := &telemetry.Config{
		ServiceName: "wemcp-gateway",
		Enabled:     gin.Mode() == gin.ReleaseMode,
	}
	otelProviders, err := telemetry.Init(ctx.Context, otelConfig)
	return otelProviders, err
}

func NewServer(ctx *cli.Context, dynamicMCPServer *service.DynamicMCPServer, otelProviders *telemetry.Providers, mcpClientService *repository.McpClientService, logger *zap.Logger) (*Server, error) {
	mcpMetrics := telemetry.NewNoopCustomMetrics()

	s := &Server{
		mcpClientService: mcpClientService,
		dynamicMCPServer: dynamicMCPServer,
		otelProviders:    otelProviders,
		metrics:          mcpMetrics,
		logger:           logger,
		ctx:              ctx,
	}

	// Set up the router after the server is fully initialized
	r, err := s.setupRouter()
	if err != nil {
		return nil, err
	}

	s.Server = &http.Server{
		Addr:    fmt.Sprintf(":%s", ctx.String("port")),
		Handler: r,
	}
	return s, nil
}

func (s *Server) setupRouter() (*http.ServeMux, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// if otel is enabled, setup prometheus metrics endpoint
	if s.otelProviders.IsEnabled() {
		// instrument gin
		r.Use(otelgin.Middleware(s.otelProviders.ServiceName()))

		// expose prometheus metrics endpoint
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	r.GET(
		"/health",
		func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		},
	)

	httpMux := http.NewServeMux()

	httpMux.Handle("/", r)

	httpMux.Handle("/mcp/{name}", s.chainMiddleware(s.dynamicMCPServer))

	return httpMux, nil
}

// Start runs the Gin server (blocking call)
func (s *Server) Start() error {
	s.logger.Info("Http server start success", zap.String("address", s.Server.Addr))
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

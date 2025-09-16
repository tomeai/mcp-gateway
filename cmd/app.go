package main

import (
	"context"
	"fmt"
	"github.com/tomeai/mcp-gateway/internal/api"
	"github.com/tomeai/mcp-gateway/internal/db"
	"github.com/tomeai/mcp-gateway/internal/service/mcpclient"
	"github.com/tomeai/mcp-gateway/internal/telemetry"
	"github.com/tomeai/mcp-gateway/utils"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	ctx      context.Context
	cancel   context.CancelFunc
	exitChan <-chan os.Signal
}

func NewApp(exitChan <-chan os.Signal) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		ctx:      ctx,
		cancel:   cancel,
		exitChan: exitChan,
	}
}

func (app *App) Run(args []string) {
	cliV2 := cli.NewApp()
	cliV2.Name = "wemcp-gateway"
	cliV2.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "port",
			Value: "8000",
		},
		&cli.StringFlag{
			Name:  "dsn",
			Value: "",
		},
	}
	cliV2.Action = func(c *cli.Context) error {
		options := []fx.Option{
			// go context
			fx.Provide(func() context.Context {
				return app.ctx
			}),
			// fx context
			fx.Provide(func() *cli.Context {
				return c
			}),
			// log
			fx.Provide(func() *zap.Logger {
				return utils.ZlogInit()
			}),
		}
		options = append(options,
			fx.Provide(api.NewOtel),
			fx.Provide(db.NewDBConnection),
			fx.Provide(mcpclient.NewMCPClientService),
			fx.Provide(api.NewServer),
			fx.Invoke(NewHttpServer),
		)
		depInj := fx.New(options...)
		if err := depInj.Start(app.ctx); err != nil {
			return err
		}

		<-app.exitChan
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := depInj.Stop(stopCtx); err != nil {
			fmt.Printf("[Fx] ERROR: Failed to stop cleanly: %v\n", err)
		}
		app.cancel()
		fmt.Printf("[Fx] Cleanly stopped\n")
		return nil
	}
	_ = cliV2.RunContext(app.ctx, args)
}

func NewHttpServer(lc fx.Lifecycle, server *api.Server, otel *telemetry.Providers, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.Start(); err != nil {
					fmt.Printf("[Fx] ERROR: Failed to start server: %v\n", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Server.Shutdown(ctx); err != nil {
				logger.Error("HTTP server Shutdown", zap.Error(err))
			}
			return otel.Shutdown(ctx)
		},
	})
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	NewApp(c).Run(os.Args)
}

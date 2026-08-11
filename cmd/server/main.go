package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yesimakar/mcp-observe-go/internal/audit"
	"github.com/yesimakar/mcp-observe-go/internal/config"
	"github.com/yesimakar/mcp-observe-go/internal/gateway"
	"github.com/yesimakar/mcp-observe-go/internal/observability"
	"github.com/yesimakar/mcp-observe-go/internal/policy"
	"github.com/yesimakar/mcp-observe-go/internal/tools"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	shutdownTelemetry, err := observability.Configure(ctx, cfg.ServiceName, cfg.OTelExporter, cfg.OTLPEndpoint)
	if err != nil {
		log.Fatalf("configure telemetry: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Printf("shutdown telemetry: %v", err)
		}
	}()

	server := gateway.NewServer(
		tools.NewRegistry(),
		policy.NewEngine(),
		audit.NewLogger(cfg.AuditPath),
	)

	httpServer := &http.Server{
		Addr:              cfg.Address,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		fmt.Printf("mcp-observe-go listening on http://localhost%s\n", cfg.Address)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}

// Pi-Star MCP (Master Control Program)
//
// Single binary that serves as the web dashboard and process supervisor
// for Pi-Star v5 amateur radio hotspots. Manages Mosquitto, MMDVMHost,
// and gateway processes; serves the dashboard UI over HTTPS; relays
// MQTT messages to browser clients via WebSocket.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/server"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/tlsutil"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	configPath := flag.String("config", "/etc/pistar-dashboard/dashboard.ini", "path to dashboard config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("pistar-dashboard", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Pi-Star dashboard starting", "version", version)

	// Step 1: Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "path", *configPath, "error", err)
		os.Exit(1)
	}
	slog.Info("config loaded", "path", *configPath)

	// Set up root context with signal-based cancellation
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Step 2: Ensure TLS certificates exist (generate self-signed if needed)
	if err := tlsutil.EnsureCerts(cfg.TLS.CertFile, cfg.TLS.KeyFile, cfg.TLS.AutoGenerate); err != nil {
		slog.Error("TLS certificate setup failed", "error", err)
		os.Exit(1)
	}
	slog.Info("TLS certificates ready", "cert", cfg.TLS.CertFile, "key", cfg.TLS.KeyFile)

	// Step 3: Start process supervisor (Mosquitto first, then MMDVM services)
	// TODO: supervisor.Start(ctx, cfg)
	slog.Info("TODO: start process supervisor")

	// Step 4: Connect MQTT client
	// TODO: mqttclient.Connect(ctx, cfg.MQTT)
	slog.Info("TODO: connect MQTT client")

	// Step 5: Discover modules
	// TODO: modules.Discover(cfg.Dashboard.ModulesDir)
	slog.Info("TODO: discover modules")

	// Step 6: Start HTTPS server
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe(ctx, cfg, content)
	}()

	slog.Info("startup sequence complete", "listen_https", cfg.Dashboard.ListenHTTPS, "listen_http", cfg.Dashboard.ListenHTTP)

	// Block until shutdown signal or server failure
	select {
	case err := <-serverErr:
		if err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received, cleaning up")
		// Wait for server graceful shutdown to finish
		if err := <-serverErr; err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}

	// TODO: Stop child processes, close connections, drain WebSocket clients
	slog.Info("shutdown complete")
}

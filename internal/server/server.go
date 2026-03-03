// Package server implements the HTTPS server, TLS configuration,
// HTTP-to-HTTPS redirect, and route registration for the Pi-Star
// dashboard.
package server

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/MW0MWZ/Pi-Star_MCP/internal/config"
	"github.com/MW0MWZ/Pi-Star_MCP/internal/hwdetect"
)

// ListenAndServe starts the HTTPS dashboard server and an HTTP redirect
// server. It blocks until ctx is cancelled, then gracefully shuts down
// both servers with a 10-second deadline.
func ListenAndServe(ctx context.Context, cfg *config.Config, content embed.FS, configPath string, devices []hwdetect.DetectedDevice, i2cDevices []hwdetect.DetectedI2CDevice) error {
	router := NewRouter(content, cfg, configPath, devices, i2cDevices)

	tlsCfg := &tls.Config{
		MinVersion:       parseTLSVersion(cfg.TLS.MinVersion),
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	httpsServer := &http.Server{
		Addr:         cfg.Dashboard.ListenHTTPS,
		Handler:      router,
		TLSConfig:    tlsCfg,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	httpServer := &http.Server{
		Addr:         cfg.Dashboard.ListenHTTP,
		Handler:      redirectHandler(cfg.Dashboard.ListenHTTPS),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 2)

	// Start HTTPS server
	go func() {
		slog.Info("HTTPS server starting", "addr", cfg.Dashboard.ListenHTTPS)
		if err := httpsServer.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTPS server: %w", err)
		}
	}()

	// Start HTTP redirect server
	go func() {
		slog.Info("HTTP redirect server starting", "addr", cfg.Dashboard.ListenHTTP)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP redirect server: %w", err)
		}
	}()

	// Wait for context cancellation or startup failure
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// Graceful shutdown
	slog.Info("shutting down HTTP servers")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpsServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTPS server shutdown error", "error", err)
	}
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP redirect server shutdown error", "error", err)
	}

	return nil
}

// redirectHandler returns an HTTP handler that redirects all requests to HTTPS.
// It is port-aware: if the HTTPS listen address uses a non-standard port,
// the redirect URL includes that port (useful for development).
func redirectHandler(httpsAddr string) http.Handler {
	_, port, _ := net.SplitHostPort(httpsAddr)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		// Strip any existing port from the Host header
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}

		target := "https://" + host
		if port != "" && port != "443" {
			target += ":" + port
		}
		target += r.URL.RequestURI()

		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}

func parseTLSVersion(v string) uint16 {
	switch v {
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

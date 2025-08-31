package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"vless-generator/internal/config"
	"vless-generator/internal/handlers"
	"vless-generator/internal/middleware"
	"vless-generator/internal/templates"
)

func main() {
	// Load configuration from command line flags
	cfg := config.LoadConfig()

	// Setup structured logging with logrus
	config.SetupLogging(cfg)

	logger := logrus.WithField("component", "main")
	logger.WithFields(logrus.Fields{
		"service": "vless-generator",
		"version": "1.0.0",
		"port":    cfg.Server.Port,
	}).Info("Starting VLESS Config Generator service")

	// Initialize template manager
	templateManager := templates.NewManager()

	// Load configuration templates
	if err := templateManager.LoadTemplates(cfg.Templates.Directory, cfg.Templates.Types); err != nil {
		logger.WithError(err).Fatal("Failed to load configuration templates")
	}

	// Update templates with configuration values
	templateManager.UpdateTemplates(
		cfg.Server.Host,
		cfg.Server.ServerPort,
		cfg.Server.WSPath,
		cfg.Network.DNSServer,
		cfg.Network.DOHServer,
		cfg.Network.TunAddress,
		cfg.Network.MixedPort,
		cfg.Network.TunMTU,
	)

	// Initialize HTTP handlers
	handler := handlers.NewHandler(templateManager)

	// Setup HTTP routes with middleware
	http.HandleFunc("/", middleware.LoggingMiddleware(handler.ConfigPageHandler))
	http.HandleFunc("/config/", middleware.LoggingMiddleware(handler.ConfigDownloadHandler))
	http.HandleFunc("/health", middleware.LoggingMiddleware(handler.HealthHandler))

	// Setup graceful shutdown
	setupGracefulShutdown(logger)

	// Start HTTP server
	serverAddr := ":" + cfg.Server.Port
	logger.WithFields(logrus.Fields{
		"address": serverAddr,
		"server":  cfg.Server.Host,
		"port":    cfg.Server.ServerPort,
		"ws_path": cfg.Server.WSPath,
	}).Info("HTTP server starting")

	logger.Info("Service endpoints available:")
	logger.Infof("  Config pages: http://localhost:%s/<type>/<uuid>", cfg.Server.Port)
	logger.Info("  Available types: neko, vless")
	logger.Infof("  Health check: http://localhost:%s/health", cfg.Server.Port)
	logger.Infof("  Config downloads: http://localhost:%s/config/<type>/<uuid>.json", cfg.Server.Port)

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		logger.WithError(err).Fatal("HTTP server failed to start")
	}
}

// setupGracefulShutdown configures graceful shutdown handling
func setupGracefulShutdown(logger *logrus.Entry) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-c
		logger.WithField("signal", sig.String()).Info("Received shutdown signal")
		logger.Info("VLESS Config Generator service shutting down gracefully")
		os.Exit(0)
	}()
}

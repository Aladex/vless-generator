package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"vless-generator/internal/config"
	"vless-generator/internal/handlers"
	"vless-generator/internal/i18n"
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

	// Initialize i18n manager
	i18nManager := i18n.NewI18n()
	if err := i18nManager.LoadTranslations("internal/i18n"); err != nil {
		logger.WithError(err).Fatal("Failed to load translations")
	}

	// Initialize template renderer
	templateRenderer := templates.NewTemplateRenderer()
	if err := templateRenderer.LoadTemplates("web/templates"); err != nil {
		logger.WithError(err).Fatal("Failed to load HTML templates")
	}

	// Initialize template manager
	templateManager := templates.NewManager()

	// Load configuration templates (without static configuration)
	if err := templateManager.LoadTemplates(cfg.Templates.Directory, cfg.Templates.Types); err != nil {
		logger.WithError(err).Fatal("Failed to load configuration templates")
	}

	// Initialize HTTP handlers
	handler := handlers.NewHandler(templateManager, templateRenderer, i18nManager)

	// Setup HTTP routes with middleware
	http.HandleFunc("/", middleware.LoggingMiddleware(handler.HomePageHandler))
	http.HandleFunc("/vless/", middleware.LoggingMiddleware(handler.ConfigPageHandler))
	http.HandleFunc("/config/", middleware.LoggingMiddleware(handler.ConfigDownloadHandler))
	http.HandleFunc("/qrcode", middleware.LoggingMiddleware(handler.QRCodeHandler))
	http.HandleFunc("/health", middleware.LoggingMiddleware(handler.HealthHandler))

	// Setup static file serving
	fs := http.FileServer(http.Dir("web/static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Setup graceful shutdown
	setupGracefulShutdown(logger)

	// Start HTTP server
	serverAddr := ":" + cfg.Server.Port
	logger.WithField("address", serverAddr).Info("HTTP server starting")

	logger.Info("Service endpoints available:")
	logger.Infof("  Home page: http://localhost:%s/", cfg.Server.Port)
	logger.Infof("  Config pages: http://localhost:%s/<type>/<uuid>?server=example.com&port=443&ws-path=/websocket&lang=ru", cfg.Server.Port)
	logger.Info("  Available types: vless")
	logger.Infof("  Health check: http://localhost:%s/health", cfg.Server.Port)
	logger.Infof("  Config downloads: http://localhost:%s/config/<type>/<uuid>.json?server=example.com", cfg.Server.Port)

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

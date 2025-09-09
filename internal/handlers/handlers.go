package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"

	"vless-generator/internal/config"
	"vless-generator/internal/i18n"
	"vless-generator/internal/templates"
	"vless-generator/internal/utils"
)

// Handler manages HTTP request handling
type Handler struct {
	templateManager  *templates.Manager
	templateRenderer *templates.TemplateRenderer
	i18n             *i18n.I18n
	logger           *logrus.Entry
}

// NewHandler creates a new handler instance
func NewHandler(templateManager *templates.Manager, templateRenderer *templates.TemplateRenderer, i18nManager *i18n.I18n) *Handler {
	return &Handler{
		templateManager:  templateManager,
		templateRenderer: templateRenderer,
		i18n:             i18nManager,
		logger:           logrus.WithField("component", "handlers"),
	}
}

// HomePageHandler handles the main page with configuration form
func (h *Handler) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	// Only handle root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Detect language from query parameter
	language := i18n.DetectLanguage(r.URL.Query().Get("lang"))

	h.logger.WithFields(logrus.Fields{
		"method":      r.Method,
		"language":    language,
		"remote_addr": r.RemoteAddr,
	}).Info("Serving home page with configuration form")

	// Get texts for the detected language
	texts := h.i18n.GetTexts(language)

	// Prepare template data
	data := templates.HomePageData{
		Title:         texts["title"],
		Language:      language,
		Texts:         texts,
		DefaultConfig: config.DefaultDynamicConfig(),
	}

	// Render template
	htmlContent, err := h.templateRenderer.RenderHomePage(data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render home page template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if _, err := fmt.Fprint(w, htmlContent); err != nil {
		h.logger.WithError(err).Error("Failed to write home page response")
	}
}

// ConfigPageHandler handles requests for configuration pages with QR codes
func (h *Handler) ConfigPageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse URL path: /<type>/<uuid>
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		h.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
		}).Warn("Invalid request path format")
		http.NotFound(w, r)
		return
	}

	configType := parts[0]
	uuid := parts[1]

	// Detect language from query parameter
	language := i18n.DetectLanguage(r.URL.Query().Get("lang"))

	// Parse dynamic configuration from query parameters
	dynamicCfg := config.ParseDynamicConfig(r.URL.Query())

	h.logger.WithFields(logrus.Fields{
		"config_type": configType,
		"uuid":        uuid,
		"language":    language,
		"server":      dynamicCfg.Server,
		"server_port": dynamicCfg.ServerPort,
		"ws_path":     dynamicCfg.WSPath,
		"remote_addr": r.RemoteAddr,
	}).Info("Generating configuration page with dynamic parameters")

	// Generate configuration with dynamic parameters
	template, err := h.templateManager.GenerateConfig(configType, uuid, dynamicCfg)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"config_type": configType,
			"uuid":        uuid,
		}).Warn("Invalid configuration type or generation failed")
		http.NotFound(w, r)
		return
	}

	// Generate VLESS URL for QR code
	vlessURL, err := utils.GenerateVlessURL(template, uuid)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"config_type": configType,
			"uuid":        uuid,
		}).Error("Failed to generate VLESS URL")
		http.Error(w, "Failed to generate configuration URL", http.StatusInternalServerError)
		return
	}

	// Generate QR code
	qr, err := qrcode.Encode(vlessURL, qrcode.Medium, 256)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"config_type": configType,
			"uuid":        uuid,
		}).Error("Failed to generate QR code")
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	// Get texts for the detected language
	texts := h.i18n.GetTexts(language)

	// Prepare query string for download links
	queryString := r.URL.RawQuery

	// Prepare template data
	data := templates.ConfigPageData{
		Title:       texts["title"],
		Language:    language,
		Texts:       texts,
		ConfigType:  strings.ToUpper(configType),
		UUID:        uuid,
		QRCode:      utils.EncodeBase64(qr),
		VlessURL:    vlessURL,
		QueryString: queryString,
	}

	// Render template
	htmlContent, err := h.templateRenderer.RenderConfigPage(data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render config page template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if _, err := fmt.Fprint(w, htmlContent); err != nil {
		h.logger.WithError(err).Error("Failed to write HTML response")
	}
}

// ConfigDownloadHandler handles JSON configuration file downloads
func (h *Handler) ConfigDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse URL path: /config/<type>/<uuid>.json
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "config" || parts[1] == "" || !strings.HasSuffix(parts[2], ".json") {
		h.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
		}).Warn("Invalid config download path format")
		http.NotFound(w, r)
		return
	}

	configType := parts[1]
	uuid := strings.TrimSuffix(parts[2], ".json")

	// Parse dynamic configuration from query parameters
	dynamicCfg := config.ParseDynamicConfig(r.URL.Query())

	h.logger.WithFields(logrus.Fields{
		"config_type": configType,
		"uuid":        uuid,
		"server":      dynamicCfg.Server,
		"server_port": dynamicCfg.ServerPort,
		"remote_addr": r.RemoteAddr,
	}).Info("Generating configuration file download with dynamic parameters")

	// Generate configuration with dynamic parameters
	cfg, err := h.templateManager.GenerateConfig(configType, uuid, dynamicCfg)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"config_type": configType,
			"uuid":        uuid,
		}).Warn("Invalid configuration type or generation failed")
		http.NotFound(w, r)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-config.json", configType))

	// Encode and send JSON
	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"config_type": configType,
			"uuid":        uuid,
		}).Error("Failed to encode configuration JSON")
		http.Error(w, "Failed to encode configuration", http.StatusInternalServerError)
		return
	}
}

// HealthHandler provides health check endpoint
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "vless-generator",
		"version":   "1.0.0",
		"templates": h.templateManager.GetTemplateTypes(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithError(err).Error("Failed to encode health response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Health check completed successfully")
}

// generateConfig creates a configuration with the specified UUID
func (h *Handler) generateConfig(template map[string]interface{}, uuid string) map[string]interface{} {
	// Deep copy the template
	cfg := utils.DeepCopyMap(template)

	// Set UUID in the first outbound (proxy)
	if outbounds, ok := cfg["outbounds"].([]interface{}); ok && len(outbounds) > 0 {
		if outbound, ok := outbounds[0].(map[string]interface{}); ok {
			if outbound["type"] == "vless" {
				outbound["uuid"] = uuid
				h.logger.WithField("uuid", uuid).Debug("UUID set in configuration")
			}
		}
	}

	return cfg
}

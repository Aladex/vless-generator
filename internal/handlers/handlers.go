package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"

	"vless-generator/internal/templates"
	"vless-generator/internal/utils"
)

// Handler manages HTTP request handling
type Handler struct {
	templateManager *templates.Manager
	logger          *logrus.Entry
}

// NewHandler creates a new handler instance
func NewHandler(templateManager *templates.Manager) *Handler {
	return &Handler{
		templateManager: templateManager,
		logger:          logrus.WithField("component", "handlers"),
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

	// Validate config type
	template, exists := h.templateManager.GetTemplate(configType)
	if !exists {
		h.logger.WithFields(logrus.Fields{
			"config_type": configType,
			"remote_addr": r.RemoteAddr,
		}).Warn("Invalid configuration type requested")
		http.NotFound(w, r)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"config_type": configType,
		"uuid":        uuid,
		"remote_addr": r.RemoteAddr,
	}).Info("Generating configuration page")

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

	// Render HTML page
	w.Header().Set("Content-Type", "text/html")
	htmlContent := h.generateHTML(configType, uuid, vlessURL, utils.EncodeBase64(qr))

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

	// Validate config type
	template, exists := h.templateManager.GetTemplate(configType)
	if !exists {
		h.logger.WithFields(logrus.Fields{
			"config_type": configType,
			"remote_addr": r.RemoteAddr,
		}).Warn("Invalid configuration type for download")
		http.NotFound(w, r)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"config_type": configType,
		"uuid":        uuid,
		"remote_addr": r.RemoteAddr,
	}).Info("Generating configuration file download")

	// Generate configuration
	cfg := h.generateConfig(template, uuid)
	if cfg == nil {
		h.logger.WithFields(logrus.Fields{
			"config_type": configType,
			"uuid":        uuid,
		}).Error("Failed to generate configuration")
		http.Error(w, "Failed to generate configuration", http.StatusInternalServerError)
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

// generateHTML creates the HTML page for configuration display
func (h *Handler) generateHTML(configType, uuid, vlessURL, qrBase64 string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>VLESS Config Generator - %s</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            text-align: center; 
            margin: 0; 
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container { 
            max-width: 600px; 
            background: white;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            padding: 40px;
        }
        .header {
            margin-bottom: 30px;
        }
        h1 {
            color: #333;
            margin: 0 0 10px 0;
            font-size: 2em;
        }
        .subtitle {
            color: #666;
            margin: 0;
        }
        .qr-section {
            margin: 30px 0;
        }
        img { 
            margin: 20px 0; 
            border: 2px solid #eee; 
            border-radius: 10px;
            max-width: 100%%;
            height: auto;
        }
        button { 
            padding: 15px 30px; 
            font-size: 16px; 
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white; 
            border: none; 
            border-radius: 25px; 
            cursor: pointer; 
            margin: 10px;
            transition: transform 0.2s, box-shadow 0.2s;
            font-weight: 600;
        }
        button:hover { 
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(0,0,0,0.2);
        }
        .info-box { 
            font-family: 'Courier New', monospace; 
            background-color: #f8f9fa; 
            padding: 15px; 
            border-radius: 8px; 
            margin: 20px 0; 
            border-left: 4px solid #667eea;
        }
        .config-type { 
            background-color: #e9ecef; 
            padding: 8px 16px; 
            border-radius: 20px; 
            display: inline-block;
            margin: 10px 0;
            font-weight: 600;
            color: #495057;
        }
        .vless-url { 
            background-color: #fff3cd; 
            padding: 15px; 
            border-radius: 8px; 
            margin: 20px 0; 
            word-break: break-all; 
            font-size: 12px;
            border-left: 4px solid #ffc107;
            text-align: left;
        }
        .instructions {
            color: #666;
            margin: 20px 0;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>VLESS Config Generator</h1>
            <p class="subtitle">Secure proxy configuration made simple</p>
        </div>
        
        <div class="config-type">Type: %s</div>
        <div class="info-box">UUID: %s</div>
        
        <div class="qr-section">
            <p class="instructions">Scan the QR code with your VLESS client to import configuration:</p>
            <img src="data:image/png;base64,%s" alt="VLESS Configuration QR Code" />
        </div>
        
        <div class="vless-url">%s</div>
        
        <div style="margin-top: 30px;">
            <a href="/config/%s/%s.json" download>
                <button>📄 Download JSON Config</button>
            </a>
        </div>
        
        <div class="instructions">
            <p><strong>Instructions:</strong></p>
            <p>1. Scan the QR code with your VLESS client (NekoBox, v2rayN, etc.)</p>
            <p>2. Or copy the VLESS URL above manually</p>
            <p>3. Alternatively, download the JSON configuration file</p>
        </div>
    </div>
</body>
</html>`, strings.ToUpper(configType), strings.ToUpper(configType), uuid, qrBase64, vlessURL, configType, uuid)
}

package templates

import (
	"embed"
	"encoding/json"
	"fmt"

	"vless-generator/internal/config"

	"github.com/sirupsen/logrus"
)

// Manager handles template loading and management
type Manager struct {
	templates map[string]map[string]interface{}
	logger    *logrus.Entry
	configFS  embed.FS
}

// NewManager creates a new template manager with embedded filesystem
func NewManager(configFS embed.FS) *Manager {
	return &Manager{
		templates: make(map[string]map[string]interface{}),
		logger:    logrus.WithField("component", "templates"),
		configFS:  configFS,
	}
}

// LoadTemplates loads embedded configuration templates from main package
func (m *Manager) LoadTemplates(types []string) error {
	m.logger.WithField("types", types).Info("Loading embedded configuration templates from templates directory")

	for _, templateType := range types {
		if err := m.loadTemplate(templateType); err != nil {
			return fmt.Errorf("failed to load template %s: %w", templateType, err)
		}
	}

	m.logger.WithField("count", len(m.templates)).Info("All templates loaded successfully")
	return nil
}

// loadTemplate loads a single embedded template file from main package
func (m *Manager) loadTemplate(templateType string) error {
	templateFile := "templates/" + templateType + ".json"

	m.logger.WithFields(logrus.Fields{
		"type": templateType,
		"file": templateFile,
	}).Debug("Loading embedded template file")

	data, err := m.configFS.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read embedded template file: %w", err)
	}

	var template map[string]interface{}
	if err := json.Unmarshal(data, &template); err != nil {
		return fmt.Errorf("failed to parse template JSON: %w", err)
	}

	m.templates[templateType] = template

	m.logger.WithField("type", templateType).Info("Template loaded successfully")
	return nil
}

// GetTemplate returns a copy of the template for the specified type
func (m *Manager) GetTemplate(templateType string) (map[string]interface{}, bool) {
	template, exists := m.templates[templateType]
	if !exists {
		return nil, false
	}

	// Return a deep copy to avoid modifying the original template
	return m.deepCopyMap(template), true
}

// GetTemplateTypes returns all available template types
func (m *Manager) GetTemplateTypes() []string {
	types := make([]string, 0, len(m.templates))
	for templateType := range m.templates {
		types = append(types, templateType)
	}
	return types
}

// GenerateConfig creates a configuration with dynamic parameters
func (m *Manager) GenerateConfig(templateType, uuid string, dynamicCfg *config.DynamicConfig) (map[string]interface{}, error) {
	template, exists := m.GetTemplate(templateType)
	if !exists {
		return nil, fmt.Errorf("template type %s not found", templateType)
	}

	// Apply dynamic configuration to the template
	m.updateTemplateWithDynamicConfig(template, dynamicCfg)

	// Set UUID in the first outbound (proxy)
	if outbounds, ok := template["outbounds"].([]interface{}); ok && len(outbounds) > 0 {
		if outbound, ok := outbounds[0].(map[string]interface{}); ok {
			if outbound["type"] == "vless" {
				outbound["uuid"] = uuid
				m.logger.WithField("uuid", uuid).Debug("UUID set in configuration")
			}
		}
	}

	return template, nil
}

// updateTemplateWithDynamicConfig updates a template with dynamic configuration values
func (m *Manager) updateTemplateWithDynamicConfig(template map[string]interface{}, dynamicCfg *config.DynamicConfig) {
	// Update server address and port in the template
	if outbounds, ok := template["outbounds"].([]interface{}); ok && len(outbounds) > 0 {
		if outbound, ok := outbounds[0].(map[string]interface{}); ok {
			outbound["server"] = dynamicCfg.Server
			outbound["server_port"] = dynamicCfg.ServerPort

			// Update WebSocket path and Host header
			if transport, ok := outbound["transport"].(map[string]interface{}); ok {
				transport["path"] = dynamicCfg.WSPath
				if headers, ok := transport["headers"].(map[string]interface{}); ok {
					headers["Host"] = dynamicCfg.Server
				}
			}

			// Update TLS server name if it exists
			if tls, ok := outbound["tls"].(map[string]interface{}); ok {
				if _, hasServerName := tls["server_name"]; hasServerName {
					tls["server_name"] = dynamicCfg.Server
				}
			}
		}
	}

	// Update DNS servers
	if dns, ok := template["dns"].(map[string]interface{}); ok {
		if servers, ok := dns["servers"].([]interface{}); ok && len(servers) >= 2 {
			if server0, ok := servers[0].(map[string]interface{}); ok {
				server0["address"] = dynamicCfg.DNSServer
			}
			if server1, ok := servers[1].(map[string]interface{}); ok {
				server1["address"] = dynamicCfg.DOHServer
			}
		}
	}

	// Update TUN address and MTU
	if inbounds, ok := template["inbounds"].([]interface{}); ok {
		if len(inbounds) > 0 {
			if inbound, ok := inbounds[0].(map[string]interface{}); ok {
				inbound["inet4_address"] = []string{dynamicCfg.TunAddress}
				inbound["mtu"] = dynamicCfg.TunMTU
			}
		}

		// Update mixed port
		if len(inbounds) > 1 {
			if inboundMixed, ok := inbounds[1].(map[string]interface{}); ok {
				inboundMixed["listen_port"] = dynamicCfg.MixedPort
			}
		}
	}
}

// deepCopyMap creates a deep copy of a map
func (m *Manager) deepCopyMap(original map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			result[key] = m.deepCopyMap(v)
		case []interface{}:
			result[key] = m.deepCopySlice(v)
		default:
			result[key] = value
		}
	}
	return result
}

// deepCopySlice creates a deep copy of a slice
func (m *Manager) deepCopySlice(original []interface{}) []interface{} {
	result := make([]interface{}, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			result[i] = m.deepCopyMap(v)
		case []interface{}:
			result[i] = m.deepCopySlice(v)
		default:
			result[i] = value
		}
	}
	return result
}

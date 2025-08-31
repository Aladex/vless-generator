package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Manager handles template loading and management
type Manager struct {
	templates map[string]map[string]interface{}
	logger    *logrus.Entry
}

// NewManager creates a new template manager
func NewManager() *Manager {
	return &Manager{
		templates: make(map[string]map[string]interface{}),
		logger:    logrus.WithField("component", "templates"),
	}
}

// LoadTemplates loads all template files from the specified directory
func (m *Manager) LoadTemplates(directory string, types []string) error {
	m.logger.WithFields(logrus.Fields{
		"directory": directory,
		"types":     types,
	}).Info("Loading configuration templates")

	for _, templateType := range types {
		if err := m.loadTemplate(directory, templateType); err != nil {
			return fmt.Errorf("failed to load template %s: %w", templateType, err)
		}
	}

	m.logger.WithField("count", len(m.templates)).Info("All templates loaded successfully")
	return nil
}

// loadTemplate loads a single template file
func (m *Manager) loadTemplate(directory, templateType string) error {
	templatePath := filepath.Join(directory, templateType+".json")

	m.logger.WithFields(logrus.Fields{
		"type": templateType,
		"path": templatePath,
	}).Debug("Loading template file")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	var template map[string]interface{}
	if err := json.Unmarshal(data, &template); err != nil {
		return fmt.Errorf("failed to parse template JSON: %w", err)
	}

	m.templates[templateType] = template

	m.logger.WithField("type", templateType).Info("Template loaded successfully")
	return nil
}

// UpdateTemplates updates all templates with the provided configuration
func (m *Manager) UpdateTemplates(server string, serverPort int, wsPath string, dnsServer, dohServer, tunAddress string, mixedPort, tunMTU int) {
	m.logger.WithFields(logrus.Fields{
		"server":      server,
		"server_port": serverPort,
		"ws_path":     wsPath,
		"dns_server":  dnsServer,
		"doh_server":  dohServer,
		"tun_address": tunAddress,
		"mixed_port":  mixedPort,
		"tun_mtu":     tunMTU,
	}).Info("Updating templates with configuration")

	for templateType, template := range m.templates {
		m.updateTemplate(template, server, serverPort, wsPath, dnsServer, dohServer, tunAddress, mixedPort, tunMTU)

		m.logger.WithField("type", templateType).Debug("Template updated")
	}

	m.logger.Info("All templates updated successfully")
}

// updateTemplate updates a single template with configuration values
func (m *Manager) updateTemplate(template map[string]interface{}, server string, serverPort int, wsPath string, dnsServer, dohServer, tunAddress string, mixedPort, tunMTU int) {
	// Update server address and port in the template
	outbound := template["outbounds"].([]interface{})[0].(map[string]interface{})
	outbound["server"] = server
	outbound["server_port"] = serverPort

	// Update WebSocket path and Host header
	transport := outbound["transport"].(map[string]interface{})
	transport["path"] = wsPath
	transport["headers"].(map[string]interface{})["Host"] = server

	// Update TLS server name if it exists
	if tls, ok := outbound["tls"].(map[string]interface{}); ok {
		if _, hasServerName := tls["server_name"]; hasServerName {
			tls["server_name"] = server
		}
	}

	// Update DNS servers
	servers := template["dns"].(map[string]interface{})["servers"].([]interface{})
	servers[0].(map[string]interface{})["address"] = dnsServer
	servers[1].(map[string]interface{})["address"] = dohServer

	// Update TUN address and MTU
	inbound := template["inbounds"].([]interface{})[0].(map[string]interface{})
	inbound["inet4_address"] = []string{tunAddress}
	inbound["mtu"] = tunMTU

	// Update mixed port
	inboundMixed := template["inbounds"].([]interface{})[1].(map[string]interface{})
	inboundMixed["listen_port"] = mixedPort

	// Update DNS rules to use the new server
	dnsRules := template["dns"].(map[string]interface{})["rules"].([]interface{})
	dnsRules[0].(map[string]interface{})["domain"] = []string{server}
}

// GetTemplate returns a template by type
func (m *Manager) GetTemplate(templateType string) (map[string]interface{}, bool) {
	template, exists := m.templates[templateType]
	return template, exists
}

// GetTemplateTypes returns all available template types
func (m *Manager) GetTemplateTypes() []string {
	types := make([]string, 0, len(m.templates))
	for templateType := range m.templates {
		types = append(types, templateType)
	}
	return types
}

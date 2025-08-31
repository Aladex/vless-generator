package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/sirupsen/logrus"
)

// DeepCopyMap creates a deep copy of a map[string]interface{}
func DeepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = DeepCopyMap(v)
		case []interface{}:
			copy[key] = deepCopySlice(v)
		default:
			copy[key] = v
		}
	}
	return copy
}

// deepCopySlice creates a deep copy of a slice
func deepCopySlice(original []interface{}) []interface{} {
	copy := make([]interface{}, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[i] = DeepCopyMap(v)
		case []interface{}:
			copy[i] = deepCopySlice(v)
		default:
			copy[i] = v
		}
	}
	return copy
}

// EncodeBase64 encodes bytes to base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// GenerateVlessURL generates a VLESS URL from template configuration
func GenerateVlessURL(template map[string]interface{}, uuid string) (string, error) {
	logger := logrus.WithFields(logrus.Fields{
		"component": "utils",
		"function":  "GenerateVlessURL",
		"uuid":      uuid,
	})

	// Extract configuration from template
	outbounds, ok := template["outbounds"].([]interface{})
	if !ok || len(outbounds) == 0 {
		return "", fmt.Errorf("invalid outbounds configuration")
	}

	outbound, ok := outbounds[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid outbound configuration")
	}

	server, ok := outbound["server"].(string)
	if !ok {
		return "", fmt.Errorf("invalid server configuration")
	}

	// Handle server_port - it could be int or float64
	var serverPort int
	switch v := outbound["server_port"].(type) {
	case int:
		serverPort = v
	case float64:
		serverPort = int(v)
	default:
		logger.WithField("type", fmt.Sprintf("%T", v)).Warn("Unexpected type for server_port, using fallback")
		serverPort = 443 // fallback
	}

	transport, ok := outbound["transport"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid transport configuration")
	}

	path, ok := transport["path"].(string)
	if !ok {
		return "", fmt.Errorf("invalid path configuration")
	}

	headers, ok := transport["headers"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid headers configuration")
	}

	host, ok := headers["Host"].(string)
	if !ok {
		return "", fmt.Errorf("invalid host configuration")
	}

	// Build VLESS URL
	vlessURL := fmt.Sprintf("vless://%s@%s:%d?type=ws&path=%s&host=%s&security=tls&fp=chrome",
		uuid, server, serverPort, path, host)

	logger.WithField("url", vlessURL).Debug("Generated VLESS URL")
	return vlessURL, nil
}

// GetScheme determines HTTP scheme from request
func GetScheme(hasTLS bool) string {
	if hasTLS {
		return "https"
	}
	return "http"
}

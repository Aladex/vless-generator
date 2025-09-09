package config

import (
	"flag"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Service   ServiceConfig
	Templates TemplatesConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string // HTTP server port only
}

// DynamicConfig holds configuration parameters from GET request
type DynamicConfig struct {
	Server     string // VLESS server address
	ServerPort int    // VLESS server port
	WSPath     string // WebSocket path
	DNSServer  string // Remote DNS server
	DOHServer  string // DNS over HTTPS server
	TunAddress string // TUN interface address
	MixedPort  int    // Mixed proxy port
	TunMTU     int    // TUN interface MTU
}

// DefaultDynamicConfig returns default values for dynamic configuration
func DefaultDynamicConfig() *DynamicConfig {
	return &DynamicConfig{
		Server:     "vless.example.com",
		ServerPort: 443,
		WSPath:     "/websocket",
		DNSServer:  "8.8.8.8",
		DOHServer:  "https://223.5.5.5/dns-query",
		TunAddress: "172.19.0.1/28",
		MixedPort:  2080,
		TunMTU:     9000,
	}
}

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	DNSServer  string
	DOHServer  string
	TunAddress string
	MixedPort  int
	TunMTU     int
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	LogLevel  string
	LogFormat string
}

// TemplatesConfig holds template-related configuration
type TemplatesConfig struct {
	Directory string
	Types     []string
}

// LoadConfig parses command-line flags and returns configuration
func LoadConfig() *Config {
	cfg := &Config{}

	// Server configuration - only port for the HTTP server
	flag.StringVar(&cfg.Server.Port, "port", "8080", "Port to run the HTTP server on")

	// Service configuration
	flag.StringVar(&cfg.Service.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.Service.LogFormat, "log-format", "json", "Log format (json, text)")

	// Templates configuration
	cfg.Templates.Directory = "templates"
	cfg.Templates.Types = []string{"vless"}

	flag.Parse()

	return cfg
}

// SetupLogging configures logrus with the specified settings
func SetupLogging(cfg *Config) {
	// Set log level
	level, err := logrus.ParseLevel(cfg.Service.LogLevel)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Set log format
	if cfg.Service.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	logrus.WithFields(logrus.Fields{
		"service": "vless-generator",
		"version": "1.0.0",
	}).Info("Logging configured successfully")
}

// ParseDynamicConfig parses dynamic configuration from URL query parameters
func ParseDynamicConfig(query url.Values) *DynamicConfig {
	config := DefaultDynamicConfig()

	if server := query.Get("server"); server != "" {
		config.Server = server
	}
	if port := query.Get("port"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.ServerPort = p
		}
	}
	if wsPath := query.Get("ws-path"); wsPath != "" {
		config.WSPath = wsPath
	}
	if dnsServer := query.Get("dns-server"); dnsServer != "" {
		config.DNSServer = dnsServer
	}
	if dohServer := query.Get("doh-server"); dohServer != "" {
		config.DOHServer = dohServer
	}
	if tunAddress := query.Get("tun-address"); tunAddress != "" {
		config.TunAddress = tunAddress
	}
	if mixedPort := query.Get("mixed-port"); mixedPort != "" {
		if mp, err := strconv.Atoi(mixedPort); err == nil {
			config.MixedPort = mp
		}
	}
	if tunMTU := query.Get("tun-mtu"); tunMTU != "" {
		if mtu, err := strconv.Atoi(tunMTU); err == nil {
			config.TunMTU = mtu
		}
	}

	return config
}

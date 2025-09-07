package config

import (
	"flag"

	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Network   NetworkConfig
	Service   ServiceConfig
	Templates TemplatesConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port       string
	Host       string
	ServerPort int
	WSPath     string
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

	// Server configuration
	flag.StringVar(&cfg.Server.Port, "port", "8080", "Port to run the server on")
	flag.StringVar(&cfg.Server.Host, "server", "vless.example.com", "VLESS server address")
	flag.IntVar(&cfg.Server.ServerPort, "server-port", 443, "VLESS server port")
	flag.StringVar(&cfg.Server.WSPath, "ws-path", "/websocket", "WebSocket path")

	// Network configuration
	flag.StringVar(&cfg.Network.DNSServer, "dns-server", "8.8.8.8", "Remote DNS server address")
	flag.StringVar(&cfg.Network.DOHServer, "doh-server", "https://223.5.5.5/dns-query", "DNS over HTTPS server")
	flag.StringVar(&cfg.Network.TunAddress, "tun-address", "172.19.0.1/28", "TUN interface address")
	flag.IntVar(&cfg.Network.MixedPort, "mixed-port", 2080, "Mixed proxy port")
	flag.IntVar(&cfg.Network.TunMTU, "tun-mtu", 9000, "TUN interface MTU")

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

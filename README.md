# VLESS Config Generator

A small, container-friendly web service that generates VLESS client configs and QR codes. Built in Go with embedded HTML/CSS and JSON templates, structured logging, and simple health checks.

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)
![Build Status](https://github.com/aladex/vless-generator/workflows/Build%20and%20Push%20Docker%20Image/badge.svg)

## Features

- VLESS config generation from a clean web UI
- On-the-fly customization via URL query parameters
- Built-in QR code endpoint and inline QR rendering
- Embedded assets (no external files at runtime)
- Structured logging with logrus
- Health check endpoint for monitoring

## Quick start

### Run locally

```bash
# Build and run
go run . -port 8080 -log-level info -log-format json

# Open UI
xdg-open http://localhost:8080/ || true

# Health
curl -s http://localhost:8080/health | jq .
```

### Docker

```bash
# Build image from this repo
docker build -t vless-generator .

# Run container
docker run --rm -p 8080:8080 vless-generator \
  -port 8080 -log-level info -log-format json

# Open UI
xdg-open http://localhost:8080/ || true
```

Note: All configuration like server, port, ws-path, etc. is provided dynamically via the UI or URL query parameters, not CLI flags.

## How it works

- The service exposes a home page with a guided wizard that builds a link to a config page.
- Config pages are served at: `/{type}/{uuid}`. Currently supported type(s): `vless`.
- Dynamic parameters are passed via query string and applied to the embedded JSON template at request time.

Example config page URL:

```
http://localhost:8080/vless/bae71742-94e0-4dd5-935f-070339819ba0?server=example.com&port=443&ws-path=/websocket&lang=en
```

## Endpoints

- GET `/` — Home page (wizard UI)
- GET `/<type>/<uuid>` — Render HTML config page with QR code (type: `vless`)
- GET `/config/<type>/<uuid>.json` — Download generated JSON configuration
- POST `/qrcode` — Generate a QR code PNG for a provided VLESS URL (form field: `url`)
- GET `/health` — Health/status JSON

Health example:

```json
{
  "status": "healthy",
  "timestamp": "2025-01-01T00:00:00Z",
  "service": "vless-generator",
  "version": "1.0.0",
  "templates": ["vless"]
}
```

## Dynamic query parameters

Pass these as query string fields to config pages or downloads:

- `server` — VLESS server hostname (e.g., example.com)
- `port` — VLESS server port (e.g., 443)
- `ws-path` — WebSocket path (e.g., /websocket)
- `dns-server` — DNS server (e.g., 8.8.8.8)
- `doh-server` — DoH server URL (e.g., https://223.5.5.5/dns-query)
- `tun-address` — TUN IPv4 address/CIDR (e.g., 172.19.0.1/28)
- `mixed-port` — Mixed inbound port (e.g., 2080)
- `tun-mtu` — TUN MTU (e.g., 9000)
- `lang` — UI language (en, ru)

Example JSON download:

```bash
curl -s "http://localhost:8080/config/vless/bae71742-94e0-4dd5-935f-070339819ba0.json?server=example.com&port=443&ws-path=/websocket" | jq .
```

## Build from source

```bash
# Install deps
go mod download

# Build binary
go build -ldflags="-w -s" -o vless-generator .

# Run it
./vless-generator -port 8080 -log-level info -log-format json
```

## Project structure (high level)

```
.
├── main.go                 # HTTP wiring and server
├── internal/
│   ├── config/             # Flags, logging, and dynamic query parsing
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # Logging middleware
│   └── templates/          # Template manager + HTML renderer
├── web/
│   ├── static/             # Embedded CSS and assets
│   └── templates/          # Embedded HTML templates
└── templates/              # Embedded JSON config templates (e.g., vless.json)
```

All HTML, CSS, and JSON templates are embedded via Go's embed; no external volumes are required at runtime.

## Kubernetes and Compose notes

- The included `docker-compose.yml` and `k8s-manifest.yaml` in this repo may show legacy CLI flags like `-server`, `-ws-path`, etc. The current version uses query parameters instead. It's safe to run the service with only logging/port flags, for example:
  - `command: ["-port", "8080", "-log-level", "info", "-log-format", "json"]`
- Then provide config details via the UI or by adding query parameters to the URLs (see examples above).

## Contributing

- Keep logs structured and user-facing text in English
- Add tests when changing public behavior
- If you add a new configuration type, place its JSON in `templates/` and include it in `internal/config/config.go` (Templates.Types).

## License

Specify a license for your fork/repo if needed.

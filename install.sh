#!/bin/bash

# Installation script for VLESS Config Generator Service

set -e

SERVICE_NAME="vless-generator"
INSTALL_DIR="/opt/vless-generator"
BINARY_NAME="vless-generator"
SERVICE_FILE="vless-generator.service"
USER="www-data"

echo "Installing VLESS Config Generator Service..."

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)"
   exit 1
fi

# Create installation directory
echo "Creating installation directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Build the Go binary
echo "Building Go binary..."
go mod tidy
go build -o "$BINARY_NAME" .

# Verify binary was created
if [[ ! -f "$BINARY_NAME" ]]; then
    echo "Error: Failed to build binary"
    exit 1
fi

# Copy binary to installation directory
echo "Installing binary to $INSTALL_DIR"
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Copy templates directory
echo "Installing templates..."
cp -r templates "$INSTALL_DIR/"

# Set ownership
echo "Setting ownership to $USER"
chown -R "$USER:$USER" "$INSTALL_DIR"

# Install systemd service
echo "Installing systemd service..."
cp "$SERVICE_FILE" "/etc/systemd/system/"

# Reload systemd and enable service
echo "Enabling and starting service..."
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

# Check service status
echo "Service status:"
systemctl status "$SERVICE_NAME" --no-pager

echo ""
echo "✅ Installation completed successfully!"
echo ""
echo "🚀 Service Information:"
echo "  Service Name: $SERVICE_NAME"
echo "  Installation Directory: $INSTALL_DIR"
echo "  Running on port: 8080 (default)"
echo ""
echo "🌐 Available Endpoints:"
echo "  Config Pages: http://localhost:8080/<type>/<uuid>"
echo "  Types: neko, vless"
echo "  Health Check: http://localhost:8080/health"
echo "  Downloads: http://localhost:8080/config/<type>/<uuid>.json"
echo ""
echo "📋 Management Commands:"
echo "  Check Status:    sudo systemctl status $SERVICE_NAME"
echo "  View Logs:       sudo journalctl -u $SERVICE_NAME -f"
echo "  Restart Service: sudo systemctl restart $SERVICE_NAME"
echo "  Stop Service:    sudo systemctl stop $SERVICE_NAME"
echo "  Start Service:   sudo systemctl start $SERVICE_NAME"
echo ""
echo "⚙️  Configuration:"
echo "  Edit: /etc/systemd/system/$SERVICE_FILE"
echo "  Then: sudo systemctl daemon-reload && sudo systemctl restart $SERVICE_NAME"
echo ""
echo "📝 Example Usage:"
echo "  Visit: http://localhost:8080/vless/your-uuid-here"
echo "  Or:    http://localhost:8080/neko/your-uuid-here"

#!/bin/bash
set -e

BIN_NAME="xampp-tui"
BIN_PATH="/usr/local/bin/$BIN_NAME"
SERVICE_PATH="/etc/systemd/system/$BIN_NAME.service"

# Verificar dependencias
if ! command -v go &> /dev/null; then
	echo "Go no está instalado. Instálalo antes de continuar."
	exit 1
fi

# Compilar binario optimizado
go build -ldflags="-s -w" -o "$BIN_NAME" ./cmd/lampp-tui/

# Copiar binario
sudo cp "$BIN_NAME" "$BIN_PATH"
sudo chmod +x "$BIN_PATH"

# Crear archivo de servicio systemd
sudo tee "$SERVICE_PATH" > /dev/null <<EOF
[Unit]
Description=XAMPP TUI Dashboard
After=network.target

[Service]
Type=simple
ExecStart=$BIN_PATH
Restart=always
User=$USER
Environment=GO_ENV=production

[Install]
WantedBy=multi-user.target
EOF

# Recargar systemd y habilitar servicio
sudo systemctl daemon-reload
sudo systemctl enable --now "$BIN_NAME"

echo "¡Instalación y servicio completados!"
echo "Puedes ver el estado con: sudo systemctl status $BIN_NAME"
echo "Para ver logs: sudo journalctl -u $BIN_NAME -f"
echo "Ejecutando la aplicación en esta terminal..."
"$BIN_PATH"
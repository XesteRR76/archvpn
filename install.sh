#!/usr/bin/env bash
set -euo pipefail

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${CYAN}"
echo "  ╔═══════════════════════════════════════════════╗"
echo "  ║        ⚡ ArchVPN Automated Installer         ║"
echo "  ║  Terminal VPN (VLESS/Reality TUN) for Arch    ║"
echo "  ╚═══════════════════════════════════════════════╝"
echo -e "${NC}"

# Check Arch Linux / pacman
if ! command -v pacman &>/dev/null; then
    echo -e "${RED}[!] Error: ArchVPN is designed specifically for Arch Linux (pacman not found).${NC}"
    exit 1
fi

echo -e "${CYAN}[*] Step 1/4: Checking and installing dependencies (sing-box, go)...${NC}"
sudo pacman -Sy --needed --noconfirm sing-box go

SINGBOX_PATH=$(command -v sing-box)
echo -e "${CYAN}[*] Step 2/4: Granting Linux capabilities to sing-box for rootless TUN...${NC}"
sudo setcap cap_net_admin,cap_net_bind_service=+ep "${SINGBOX_PATH}"

echo -e "${CYAN}[*] Step 3/4: Compiling archvpn binary from source...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

mkdir -p bin
go build -ldflags="-s -w" -o bin/archvpn ./cmd/archvpn

echo -e "${CYAN}[*] Step 4/4: Installing binary to /usr/local/bin/archvpn...${NC}"
sudo cp bin/archvpn /usr/local/bin/archvpn
sudo chmod 755 /usr/local/bin/archvpn

# Prepare config directory
mkdir -p "${HOME}/.config/archvpn"

echo -e "${GREEN}"
echo "  [✓] Installation finished successfully!"
echo "  [✓] Binary installed: /usr/local/bin/archvpn"
echo -e "${NC}"
echo -e "${YELLOW}Quick Start:${NC}"
echo "  1. Add your subscription:"
echo "     archvpn add-sub \"https://your-subscription-url\""
echo ""
echo "  2. Launch interactive TUI dashboard:"
echo "     archvpn"
echo ""
echo "  3. Or connect via CLI:"
echo "     archvpn connect"
echo ""

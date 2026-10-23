#!/usr/bin/env bash
set -euo pipefail

echo "[*] Checking and installing sing-box on Arch Linux..."
if ! command -v sing-box &>/dev/null; then
    echo "[*] Installing sing-box via pacman..."
    sudo pacman -S --needed --noconfirm sing-box
fi

SINGBOX_PATH=$(command -v sing-box)
echo "[*] sing-box located at: ${SINGBOX_PATH}"
echo "[*] Applying network capabilities (cap_net_admin, cap_net_bind_service)..."
sudo setcap cap_net_admin,cap_net_bind_service=+ep "${SINGBOX_PATH}"

echo "[*] Verifying sing-box version:"
sing-box version | head -n 2
echo "[+] Environment setup completed successfully!"

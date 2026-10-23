# ⚡ ArchVPN

<p align="center">
  <b>Высокопроизводительный терминальный VPN-клиент нового поколения для Arch Linux.</b><br>
  Ядро <code>sing-box 1.14+</code> в режиме <b>системного TUN</b> с интерактивным TUI на <code>Charm Bubbletea</code> и скриптовым CLI.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Arch_Linux-1793D1?style=for-the-badge&logo=arch-linux&logoColor=white" />
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/sing--box-1.14-blue?style=for-the-badge" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" />
</p>

---

## ✨ Ключевые фичи

- 🛡 **Системный TUN-интерфейс (`tun0`)**: Никаких локальных SOCKS/HTTP-прокси в браузерах. Весь трафик системы (игры, Telegram, Discord, pacman, curl, docker) автоматически направляется в зашифрованный туннель на уровне ядра Linux.
- 🚀 **Поддержка современных протоколов обхода блокировок**:
  - **VLESS** + **XTLS-Reality** (с `pbk`, `sid`, `spx`)
  - **XTLS-Vision flow** (`xtls-rprx-vision`)
  - **XHTTP / HTTPUpgrade**, **gRPC**, **WebSocket**, **TCP**
- ⚡ **Параллельный бенчмарк задержки**: Concurrent worker pool замеряет реальный TCP/TLS handshake пинг до 100+ серверов за 1-2 секунды.
- 🎨 **Интерактивный Charm TUI**: Эстетичный темный интерфейс (Bubbletea + Lipgloss), живой поиск по нажатию `/`, флаги стран (`🇳🇱`, `🇩🇪`, `🇷🇺`, `🇫🇮`, `🇪🇪`, `🇨🇭`...) и цветовая градация пинга.
- ⌨️ **Полноценный CLI-интерфейс**: Быстрые команды `connect`, `disconnect`, `status`, `ping`, `list`, `update` для хоткеев Hyprland/i3 и Waybar.
- 📊 **Готовая интеграция с Waybar / Polybar**: Нативный флаг `archvpn status --json`.
- 🔄 **Безопасное отключение (No Network Leak)**: Автоматическая очистка маршрутов, интерфейса `tun0` и восстановление системного DNS при выходе.

---

## 🖥 Интерфейс (TUI)

```text
  ⚡ ARCH-VPN  ● CONNECTED: 🇩🇪 Мост через Россию (00:04:12)

  ID       COUNTRY      REMARK                                 TYPE       LATENCY
  ----------------------------------------------------------------------------------
  ❯ 7987b4bc 🇩🇪 DE        🇩🇪 Whitelist                           xhttp       79 ms
    8d70cacf 🇩🇪 DE        🇩🇪 Мост через Россию                   xhttp      124 ms
    7df012da 🇷🇺 RU        🇷🇺 Россия [Ютуб без рекламы]           tcp        154 ms
    49e200ac 🇫🇮 FI        🇫🇮 LTE #3 (1x)                         xhttp      168 ms
    822b3f92 🇷🇺 RU        🇷🇺 LTE #1 (1x)                         tcp        175 ms
    58ddb2ad 🇨🇭 CH        🇨🇭 LTE #1 (1x)                         tcp        189 ms
    8ad50bae 🇩🇪 DE        🇩🇪 LTE #1 (1x)                         tcp        197 ms

  [↑/↓] Navigate  [Enter/Space] Connect  [d] Disconnect  [p] Ping All  [u] Update Subs  [/] Filter  [q] Quit
```

---

## 📦 Быстрая установка

### Способ 1: Автоматический скрипт установки (рекомендуется)

```bash
git clone https://github.com/XesteRR76/archvpn.git
cd archvpn
./install.sh
```

Скрипт автоматически:
1. Установит `sing-box` и `go` из официальных репозиториев Arch Linux (`pacman`).
2. Настроит Linux capabilities (`cap_net_admin,cap_net_bind_service=+ep`) для создания TUN-интерфейса без постоянных запросов `sudo`.
3. Скомпилирует и установит бинарник в `/usr/local/bin/archvpn`.

---

### Способ 2: Ручная сборка

```bash
# 1. Установка зависимостей
sudo pacman -Sy --needed sing-box go

# 2. Выдача сетевых прав ядру sing-box
sudo setcap cap_net_admin,cap_net_bind_service=+ep $(which sing-box)

# 3. Сборка бинарника
git clone https://github.com/XesteRR76/archvpn.git
cd archvpn
go build -ldflags="-s -w" -o bin/archvpn ./cmd/archvpn

# 4. Установка в систему
sudo cp bin/archvpn /usr/local/bin/archvpn
```

---

## 🚀 Использование

### 1. Добавление подписки

Клиент поддерживает Base64-подписки и списки `vless://` ссылок:

```bash
archvpn add-sub "https://your-subscription-provider.com/sub/token"
```

*После добавления клиент автоматически загрузит и распарсит все серверы.*

---

### 2. Интерактивный TUI-дашборд

Запуск графического интерфейса в терминале:

```bash
archvpn
```

#### Горячие клавиши:
| Клавиша | Действие |
| :--- | :--- |
| `↑` / `↓` или `k` / `j` | Навигация по списку серверов |
| `Enter` или `Space` | Подключиться к выбранному серверу |
| `d` | Отключить активный туннель |
| `p` | Запустить параллельный замер пинга всех серверов |
| `u` | Обновить подписки из сети |
| `/` | Поиск и фильтрация по названию, стране или типу (Esc — сброс) |
| `q` или `Ctrl+C` | Выход (с корректным отключением туннеля) |

---

### 3. Быстрые CLI-команды

```bash
# Подключиться к последнему выбранному или первому серверу:
archvpn connect

# Подключиться к серверу по номеру из списка:
archvpn connect 1

# Подключиться по названию страны или части имени:
archvpn connect "Германия"
archvpn connect "Нидерланды"
archvpn connect "Москва"

# Отключить VPN:
archvpn disconnect

# Посмотреть текущий статус подключения:
archvpn status

# Таблица всех доступных серверов с задержками:
archvpn list

# Запустить параллельный пинг всех серверов:
archvpn ping

# Обновить подписки:
archvpn update
```

---

## 📊 Интеграция со статус-барами (Waybar / Polybar)

`archvpn status --json` отдает форматированный JSON, идеально подходящий для Waybar.

### Настройка Waybar (`~/.config/waybar/config.jsonc`):

```jsonc
"custom/vpn": {
    "format": "{}",
    "return-type": "json",
    "interval": 2,
    "exec": "archvpn status --json",
    "on-click": "archvpn disconnect",
    "on-click-right": "alacritty -e archvpn"
}
```

### Стили (`~/.config/waybar/style.css`):

```css
#custom-vpn.connected {
    color: #10b981;
    font-weight: bold;
}

#custom-vpn.disconnected {
    color: #64748b;
}
```

---

## ⚙️ Конфигурация

Конфигурационный файл создается автоматически по пути `~/.config/archvpn/config.json`:

```json
{
  "subscription_urls": [
    "https://your-subscription-url.com/sub/token"
  ],
  "tun_interface": "tun0",
  "dns_servers": [
    "1.1.1.1",
    "8.8.8.8"
  ]
}
```

---

## 🗑 Удаление

```bash
sudo rm -f /usr/local/bin/archvpn
rm -rf ~/.config/archvpn
```

---

## 📄 Лицензия

Распространяется под лицензией MIT. Подробности в файле [LICENSE](LICENSE).

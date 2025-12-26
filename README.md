<div align="center">
  <img src="assets/golazo-logo.png" alt="Golazo demo" width="150">
  <h1>Golazo</h1>
</div>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xjuanma/golazo)](https://goreportcard.com/report/github.com/0xjuanma/golazo)
[![GitHub Release](https://img.shields.io/github/v/release/0xjuanma/golazo)](https://github.com/0xjuanma/golazo/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/0xjuanma/golazo/build.yml)](https://github.com/0xjuanma/golazo/actions/workflows/build.yml)

A minimalist terminal user interface (TUI) for following football matches in real-time. Get live match updates, finished match statistics, and minute-by-minute events directly in your terminal.
</div>

<div align="center">
  <img src="assets/golazo-ss-v0.4.0.png" alt="Golazo screenshot" width="600">
</div>

## Installation/Update

> [!IMPORTANT]
> Tool is in active development.

### Using the install script (recommended)

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershellq
irm https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.ps1 | iex
```

### Build from source

```bash
git clone https://github.com/0xjuanma/golazo.git
cd golazo
go build 
./golazo
```

## Usage

Run the application:
```bash
golazo
```

## Supported Leagues

Many leagues and competitions across Europe, South America, North America, Middle East, and more. [View full list](docs/SUPPORTED_LEAGUES.md)

Customize your leagues and competitions preferences in the **Settings** menu.

## Notification Setup

Goal notifications require one-time setup depending on your operating system.

### macOS

Notifications use AppleScript, which requires enabling notifications for Script Editor:

1. Open **Script Editor** (`/Applications/Utilities/Script Editor.app`)
2. Paste and run: `display notification "test" with title "test"`
3. Open **System Settings → Notifications → Script Editor**
4. Enable/Allow notifications and set alert style to "Banners"

### Linux

Notifications require `libnotify`. Install if not present:

```bash
# Debian/Ubuntu
sudo apt install libnotify-bin

# Fedora
sudo dnf install libnotify

# Arch
sudo pacman -S libnotify
```

### Windows

Notifications should work out-of-box on Windows 10/11.

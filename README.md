<div align="center">
  <img src="assets/golazo-logo.png" alt="Golazo demo" width="150">
  <h1>Golazo</h1>
</div>

<div align="center">

[![GitHub Stars](https://img.shields.io/github/stars/0xjuanma/golazo?style=social)](https://github.com/0xjuanma/golazo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xjuanma/golazo)](https://goreportcard.com/report/github.com/0xjuanma/golazo)
[![GitHub Release](https://img.shields.io/github/v/release/0xjuanma/golazo)](https://github.com/0xjuanma/golazo/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/0xjuanma/golazo/build.yml)](https://github.com/0xjuanma/golazo/actions/workflows/build.yml)

[![GitHub Downloads](https://img.shields.io/github/downloads/0xjuanma/golazo/total)](https://github.com/0xjuanma/golazo/releases)
![macOS](https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black)
![Windows](https://img.shields.io/badge/Windows-0078D6?logo=windows&logoColor=white)

A minimalist terminal user interface (TUI) for following football (soccer) matches in real-time. Get live match updates, finished match statistics, and minute-by-minute events directly in your terminal.

Golazo also ships a JSON CLI for agents and scripts (`golazo live`, `finished`, `match`, `leagues`, `capabilities`) — see **[CLI / Agent Mode](docs/CLI.md)**.

Golazo was created for those moments when you can't stream or watch matches live. It gives you a handy, non-intrusive, and minimalist way to keep up with your favourite football leagues.

*Perfect for developers and terminal enthusiasts who want match updates without leaving their workflow.*
</div>

> [!NOTE]
> If you enjoy Golazo, give it a star and share it with your friends. That helps others find it and keeps the project going!

<div align="center">
  <img src="assets/golazo-demo-v0.18.0.gif" alt="Golazo demo" width="800">
</div>

<div align="center">

**Quick Install:** `brew install 0xjuanma/tap/golazo` · [Other options](#installation--update)

</div>

## Features

- **Live Match Tracking**: Timeline & Real-time updates for goals, cards, and substitutions with automatic polling
- **Finished Matches**: View results from today, last 3 days, or last 5 days
- **Match Statistics & Details**: Possession, shots, passes, standings, formations with player ratings, and more in focused dialogs
- **Official Highlights & Replay Links**: Clickable links for official highlights and instant goal replays
- **Goal Notifications**: Desktop notifications for goals as they happen
- **65+ Leagues**: Organized by region (Europe, Americas, Global) with tab navigation in Settings
- **JSON CLI for agents**: `golazo live`, `finished`, `match`, `leagues`, `capabilities` — structured output, typed error codes, exit code map. See [docs/CLI.md](docs/CLI.md).

## Installation & Update

**Self-update:** Run `golazo --update` anytime to get the latest version.

### Homebrew

```bash
# Install
brew install 0xjuanma/tap/golazo

# Update
brew upgrade 0xjuanma/tap/golazo
```

### Install script

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
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

**Navigation:** `↑`/`↓` or `j`/`k` to move, `Enter` to select, `/` to filter, `Tab` to focus view, `Esc` to go back, `q` to quit.

## CLI / Agent Mode

For scripts and agentic tools (Claude Code, Codex, MCP servers), Golazo exposes JSON subcommands:

```bash
golazo capabilities | jq .                       # self-discover the contract
golazo live                                       # live matches right now
golazo finished --include-upcoming                # today's full slate
golazo finished --days 3                          # last 3 days
golazo match 2001 --mock                          # full match details (best-effort against real IDs; reliable with --mock)
golazo leagues --all                              # every supported league
```

Full contract — JSON envelope, error codes, exit codes, retry policy, schema, jq recipes — in **[docs/CLI.md](docs/CLI.md)**.

## Docs

- [Supported Leagues](docs/SUPPORTED_LEAGUES.md): Full list of available leagues and competitions, customize your preferences in the **Settings** menu.
- [Notifications](docs/NOTIFICATIONS.md): Desktop notification setup and configuration
- [CLI / Agent Mode](docs/CLI.md): JSON subcommands for agents and scripts (`golazo live`, `finished`, `match`, `leagues`)

---

<div align="center">

**Built with** [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss) & [Bubbles](https://github.com/charmbracelet/bubbles) by [Charm](https://charm.sh)

Author: [@0xjuanma](https://github.com/0xjuanma)

</div>

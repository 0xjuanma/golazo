# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Dialog Overlay System** - New reusable dialog component for displaying overlay content on top of views(when right panel focused) 
- **League Standings Dialog** - Press `s` to view current league standings with highlighted match teams
- **Match Formations Dialog** - Press `f` to view both teams' formations and starting lineups with player ratings
- **Full Statistics Dialog** - Press `x` to view full match statistics

### Changed
- **Unified Header Design** - All panel titles now use consistent compact header style with gradient text and diagonal fill pattern
- **Visual Overhaul** - Refreshed main menu logo and updated styling across views

### Fixed

## [0.17.0] - 2026-01-24

### Added

### Changed
- **Smart Update Detection** - The `--update` command now automatically detects whether golazo was installed via Homebrew or install script and uses the appropriate update method
- **Unified Match Details Rendering** - Consolidated live and finished match views into a single rendering system with consistent styling

### Fixed
- **Finished Matches Hints** - Fixed missing keyboard hints in finished matches view and added tab focus indicator

## [0.16.0] - 2026-01-22

### Added
- **Homebrew Support** - Install with `brew install 0xjuanma/tap/golazo`

### Changed

### Fixed
- **Light Terminal Support** - Colors now automatically adapt to light terminal themes for improved visibility

## [0.15.0] - 2026-01-14

### Added
- **Stoppage Time Display** - Goals in stoppage time now display properly (e.g., "45+2'")
- **More Leagues Supported** - Added Gaucho Brasilian competition and multiple Portuguese leagues and competitions (Thanks @felipeolibon and @rmscoelho!)
- **Official Match Highlights** - Finished matches now display clickable links to official highlight videos when available!
- **Penalty Shootout Results** - Finished matches now display penalty scores when matches went to shootouts
- **New Contributing Docs** - Added documentation for contributors, including contributing guidelines and process

### Changed
- **Go Version** - Updated minimum Go version 1.25

### Fixed

## [0.14.0] - 2026-01-10

### Added
- **South American Leagues** - Added support for Chilean Primera Division, Peruvian Liga 1, Ecuadorian Serie A, and Uruguayan Primera Division

### Changed

### Fixed
- **Duplicate Matches** - Fixed duplicate matches appearing in the finished matches list by adding deduplication by match ID
- **League IDs** - Corrected a few league IDs that were outdated

## [0.13.1] - 2026-01-08

### Added

### Changed

### Fixed
- **Enhanced Version Management** - Fixed version comparison logic and organized version functionality into dedicated package

## [0.13.0] - 2026-01-08

### Added
- **Regional League Settings Revamp** - Settings now organizes leagues by regions with tab navigation
- **New Leagues & Competitions** - Added Ukrainian Premier League, Russian Premier League, Chinese Super League, Qatar Stars League, Premier Soccer League (South Africa), Botola Pro (Morocco), Supercopa de España, FIFA Club World Cup, UEFA Nations League, Club Friendlies, and International Friendlies support
- **Scrollable Match Details** - Goals, Cards, and Statistics sections in finished matches view are now scrollable with Tab focus + arrow keys

### Changed

### Fixed
- **Stale Cached Banner** – Resolved incorrect banner persistence when using stale cached versions
- **Settings Filter Cursor Position** – Fixed filter cursor shift when filtering leagues in settings view by rendering checkbox independently

## [0.12.0] - 2026-01-05

### Added
- **Debug Mode** - New `--debug` CLI flag with automatic log rotation, UI indicators, and comprehensive API failure logging
- **Version Update Banner** - Automatic version checking with banner notifications when new Golazo versions are available

### Changed
- **Cache TTL Management** - Improved caching with separate TTL for successful links (7 days) and failures (5 minutes)

### Fixed
- **Goal Link Processing & Cache Expiration Logic** - Optimized goal processing logic and NOT_FOUND entries now properly expire and allow retry after 5 minutes
- **Live Goal-replay Link** - Resolved inconsistent rendering of goal links in live view

## [0.11.0] - 2026-01-03

### Added

### Changed
- **Goal Link Indicator** - Replaced 📺 emoji with [▶REPLAY] text indicator for better terminal compatibility
- **Goal Link Alignment** - Positioned replay links between player name and goal symbol for proper home/away expansion
- **Goal Display** - Removed assist information from goal events, showing only the scorer's name

### Fixed
- **Goal Link Cache Logic** - Improved caching behavior for goal replay links and fixed cache expiration logic for not-found

## [0.10.0] - 2026-01-03

### Added
- **Embedded Goal Replay Links** - Goal events now display clickable 📺 indicators that link to replay videos
- **Nix Flake support** - Added initial Nix flake for reproducible builds and development (Thanks @jcmuller)
- **13 New Leagues** - J. League (Japan), K League 1 (South Korea), Indian Super League, A-League (Australia), Egyptian Premier League, Brasileirão Série B, Copa Sudamericana, UEFA Conference League, EFL League One, EFL League Two, Allsvenskan (Sweden), Superligaen (Denmark), Super League 1 (Greece)

### Changed

### Fixed
- **Austrian Bundesliga ID** - Corrected league ID from 109 to 38
- **Duplicate "No items" message** - Resolve double "No items" message when no matches found (Thanks @neomantra)

## [0.9.0] - 2026-01-01

### Added
- **New Leagues** - Add Colombian division A & B leagues, Ireland Premier & First Division (Thanks @jcmuller & @ryandeering)

### Changed
- **Center-Aligned Event Timeline** - Match events now display with centered time, home events expand left, away events expand right

### Fixed
- **Finished Matches Navigation** - H/left & L/right arrow keys now correctly cycle timeframe

## [0.8.0] - 2025-12-31

### Added

### Changed
- **Upcoming Matches in Live View** - Today's upcoming matches now display at the bottom of the Live View instead of the Finished Matches view

### Fixed
- **Windows Self-Update** - Fixed `--update` failing when golazo is already running
- **Small Terminal Layout Overflow** - Fixed panel layout corruption when terminal window is too small to display all content
- **Linux Cache Location** - Empty results cache now uses correct XDG config directory (`~/.config/golazo`)

## [0.7.0] - 2025-12-28

### Added
- **Women's Leagues** - 10 new leagues: WSL, Liga F, Frauen-Bundesliga, Serie A Femminile, Première Ligue Féminine, NWSL, Women's UCL, UEFA Women's Euro, Women's DFB Pokal, Women's World Cup (Thanks @fkr!)
- **Notification Icon** - Goal notifications now display the golazo logo on Linux and Windows

### Changed
- **Linux config location** - Now follows XDG spec at `~/.config/golazo`

  > [!NOTE]
  > **Existing Linux users, choose one:**
  > - **Keep your settings**: `mv ~/.golazo ~/.config/golazo`
  > - **Start fresh**: `rm -rf ~/.golazo` (old location will be ignored)

### Fixed
- **Windows Rendering** - Fixed layout shift issue when navigating between matches on Windows Terminal

## [0.6.0] - 2025-12-26

### Added
- **Goal Notifications** - Desktop notifications and terminal beep for new goals in live matches using score-based detection (macOS, Linux, Windows)
- **New CLI Flags** - Added `--version/-v` to display version info and `--update/-u` to self-update to latest release

### Changed
- **Poll Spinner Duration** - Increased "Updating..." spinner display time to 1 second for better visibility

### Fixed
- **Card Colors in All Events** - Yellow and red cards now display proper colors (yellow/red) instead of cyan in the FT view's All Events section
- **Live Match Polling** - Poll refreshes now bypass cache to ensure fresh data every 90 seconds
- **Substitution Display** - Fixed inverted player order & colour coding in substitutions

## [0.5.0] - 2025-12-25

### Added
- **More Leagues & International Competitions** - EFL Championship, FA Cup, DFB Pokal, Coppa Italia, Coupe de France, Saudi Pro League, Africa Cup of Nations

### Changed
- **Settings UI Revamp** - League selection now uses scrollable list with fuzzy filtering (type `/` to search)

### Fixed

## [0.4.0] - 2025-12-24

### Added
- **Windows Support** - Added Windows builds (amd64, arm64) and PowerShell install script
- **10 New Leagues** - Eredivisie, Primeira Liga, Belgian Pro League, Scottish Premiership, Süper Lig, Swiss Super League, Austrian Bundesliga, Ekstraklasa, Copa del Rey, Liga MX

### Changed
- **Cards Section Redesign** - Cards now display detailed list with player name, minute, and team instead of just counts
- **Default Leagues** - When no leagues are selected in Settings, app now defaults to Premier League, La Liga, and Champions League (instead of all 24 leagues) for faster performance

### Fixed

## [0.3.0] - 2025-12-23

### Added
- **League Selection** - New settings customization to select and persist league preferences
- **Result List Filtering** - New / filtering command for all result lists

### Changed

### Fixed

## [0.2.0] - 2025-12-22

### Added
- **Polling Spinner** - Small gradient random spinner shows when live match data is being polled
- **Kick-off Time** - Live matches now display kick-off time (KO) in the match list

### Changed
- **Event Styling** - Minimal styling added to live events to clearly denote each type
- **Live View Layout** - Reordered match info: minute/league, teams, then large score display
- **Large Score Display** - Score now rendered in prominent block-style digits for visibility

### Fixed
- **Live Events Order** - Events now sorted by time (descending) with proper uniqueness
- **Match Navigation** - Spinner correctly resets when switching between live matches
- **List Item Height** - Match list items now properly display 3 lines to show KO time

## [0.1.0] - 2025-12-19

### Added
- Initial public release
- Live match tracking with real-time updates
- Match details view with events and statistics
- Several major footbal leagues supported
- Beautiful TUI with neon-styled interface
- FotMob API integration for match data
- Cross-platform support (macOS, Linux)


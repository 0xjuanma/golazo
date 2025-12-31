# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

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


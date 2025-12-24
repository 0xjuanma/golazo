# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Windows Support** - Added Windows builds (amd64, arm64) and PowerShell install script

### Changed

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


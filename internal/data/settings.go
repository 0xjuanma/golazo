package data

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const settingsFileName = "settings.yaml"

// LeagueInfo contains league metadata for display purposes.
type LeagueInfo struct {
	ID      int
	Name    string
	Country string
}

// AllSupportedLeagues contains all leagues that Golazo supports.
// This is the source of truth for available leagues.
var AllSupportedLeagues = []LeagueInfo{
	// Top 5 European Leagues
	{ID: 47, Name: "Premier League", Country: "England"},
	{ID: 87, Name: "La Liga", Country: "Spain"},
	{ID: 54, Name: "Bundesliga", Country: "Germany"},
	{ID: 55, Name: "Serie A", Country: "Italy"},
	{ID: 53, Name: "Ligue 1", Country: "France"},
	// Top 5 Women's Leagues
	{ID: 9227, Name: "Women's Super League", Country: "England"},
	{ID: 9907, Name: "Liga F", Country: "Spain"},
	{ID: 9676, Name: "Frauen-Bundesliga", Country: "Germany"},
	{ID: 10178, Name: "Serie A Femminile", Country: "Italy"},
	{ID: 9667, Name: "Première Ligue Féminine", Country: "France"},
	// Second Tier European Leagues
	{ID: 48, Name: "EFL Championship", Country: "England"},
	{ID: 57, Name: "Eredivisie", Country: "Netherlands"},
	{ID: 61, Name: "Primeira Liga", Country: "Portugal"},
	{ID: 114, Name: "Belgian Pro League", Country: "Belgium"},
	{ID: 64, Name: "Scottish Premiership", Country: "Scotland"},
	{ID: 71, Name: "Süper Lig", Country: "Turkey"},
	{ID: 66, Name: "Swiss Super League", Country: "Switzerland"},
	{ID: 109, Name: "Austrian Bundesliga", Country: "Austria"},
	{ID: 52, Name: "Ekstraklasa", Country: "Poland"},
	// Ireland
	{ID: 126, Name: "League of Ireland Premier Division", Country: "Ireland"},
	{ID: 218, Name: "League of Ireland First Division", Country: "Ireland"},
	// European Competitions
	{ID: 42, Name: "UEFA Champions League", Country: "Europe"},
	{ID: 9375, Name: "Women's UEFA Champions League", Country: "Europe"},
	{ID: 73, Name: "UEFA Europa League", Country: "Europe"},
	{ID: 50, Name: "UEFA Euro", Country: "Europe"},
	{ID: 292, Name: "UEFA Women's Euro", Country: "Europe"},
	// Domestic Cups
	{ID: 138, Name: "Copa del Rey", Country: "Spain"},
	{ID: 132, Name: "FA Cup", Country: "England"},
	{ID: 209, Name: "DFB Pokal", Country: "Germany"},
	{ID: 10650, Name: "Women's DFB Pokal", Country: "Germany"},
	{ID: 141, Name: "Coppa Italia", Country: "Italy"},
	{ID: 134, Name: "Coupe de France", Country: "France"},
	// South America
	{ID: 268, Name: "Brasileirão Série A", Country: "Brazil"},
	{ID: 112, Name: "Liga Profesional", Country: "Argentina"},
	{ID: 14, Name: "Copa Libertadores", Country: "South America"},
	{ID: 44, Name: "Copa America", Country: "South America"},
	// North America
	{ID: 130, Name: "MLS", Country: "USA"},
	{ID: 9134, Name: "NWSL", Country: "USA"},
	{ID: 230, Name: "Liga MX", Country: "Mexico"},
	// Middle East
	{ID: 536, Name: "Saudi Pro League", Country: "Saudi Arabia"},
	// International
	{ID: 77, Name: "FIFA World Cup", Country: "International"},
	{ID: 76, Name: "Women's FIFA World Cup", Country: "International"},
	{ID: 289, Name: "Africa Cup of Nations", Country: "Africa"},
}

// Settings represents user preferences stored in settings.yaml.
type Settings struct {
	// SelectedLeagues contains the IDs of leagues the user wants to follow.
	// If empty, all supported leagues are used.
	SelectedLeagues []int `yaml:"selected_leagues"`
}

// SettingsPath returns the path to the settings file.
func SettingsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, settingsFileName), nil
}

// LoadSettings reads settings from the settings.yaml file.
// Returns default settings (empty selection = all leagues) if file doesn't exist.
func LoadSettings() (*Settings, error) {
	path, err := SettingsPath()
	if err != nil {
		return &Settings{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No settings file - return empty settings (will use all leagues)
			return &Settings{}, nil
		}
		return &Settings{}, err
	}

	var settings Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		// Invalid YAML - return empty settings
		return &Settings{}, nil
	}

	return &settings, nil
}

// SaveSettings writes settings to the settings.yaml file.
func SaveSettings(settings *Settings) error {
	path, err := SettingsPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// DefaultLeagueIDs contains the default leagues used when no selection is made.
// These are the most popular leagues for efficient API usage.
var DefaultLeagueIDs = []int{
	47, // Premier League
	87, // La Liga
	42, // UEFA Champions League
}

// GetActiveLeagueIDs returns the league IDs that should be used for API calls.
// If no leagues are selected in settings, returns the default leagues (not all).
func GetActiveLeagueIDs() []int {
	settings, err := LoadSettings()
	if err != nil || len(settings.SelectedLeagues) == 0 {
		// Return default leagues for efficient API usage
		return DefaultLeagueIDs
	}

	return settings.SelectedLeagues
}

// GetAllLeagueIDs returns all supported league IDs (used as fallback).
func GetAllLeagueIDs() []int {
	ids := make([]int, len(AllSupportedLeagues))
	for i, league := range AllSupportedLeagues {
		ids[i] = league.ID
	}
	return ids
}

// IsLeagueSelected checks if a league ID is in the selected list.
func (s *Settings) IsLeagueSelected(leagueID int) bool {
	for _, id := range s.SelectedLeagues {
		if id == leagueID {
			return true
		}
	}
	return false
}

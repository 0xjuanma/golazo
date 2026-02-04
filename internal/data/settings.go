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

// Region constants for organizing leagues
const (
	RegionEurope  = "Europe"
	RegionAmerica = "Americas"
	RegionGlobal  = "Global"
)

// AllSupportedLeagues contains all leagues that Golazo supports organized by region.
// This is the source of truth for available leagues.
var AllSupportedLeagues = map[string][]LeagueInfo{
	RegionEurope: {
		// Top 5 European Leagues
		{ID: 47, Name: "Premier League", Country: "England"},
		{ID: 87, Name: "La Liga", Country: "Spain"},
		{ID: 54, Name: "Bundesliga", Country: "Germany"},
		{ID: 512, Name: "Regionalliga", Country: "Germany"},
		{ID: 55, Name: "Serie A", Country: "Italy"},
		{ID: 53, Name: "Ligue 1", Country: "France"},
		// Top 5 Women's Leagues
		{ID: 9227, Name: "Women's Super League", Country: "England"},
		{ID: 9907, Name: "Liga F", Country: "Spain"},
		{ID: 9676, Name: "Frauen-Bundesliga", Country: "Germany"},
		{ID: 10178, Name: "Serie A Femminile", Country: "Italy"},
		{ID: 9667, Name: "Première Ligue Féminine", Country: "France"},
		// Other European Leagues
		{ID: 67, Name: "Allsvenskan", Country: "Sweden"},
		{ID: 38, Name: "Austrian Bundesliga", Country: "Austria"},
		{ID: 40, Name: "Belgian First Division", Country: "Belgium"},
		{ID: 48, Name: "EFL Championship", Country: "England"},
		{ID: 108, Name: "EFL League One", Country: "England"},
		{ID: 109, Name: "EFL League Two", Country: "England"},
		{ID: 196, Name: "Ekstraklasa", Country: "Poland"},
		{ID: 57, Name: "Eredivisie", Country: "Netherlands"},
		{ID: 218, Name: "League of Ireland First Division", Country: "Ireland"},
		{ID: 126, Name: "League of Ireland Premier Division", Country: "Ireland"},
		{ID: 61, Name: "Primeira Liga", Country: "Portugal"},
		{ID: 10215, Name: "Primeira Liga Qualification", Country: "Portugal"},
		{ID: 185, Name: "Liga Portugal 2", Country: "Portugal"},
		{ID: 9668, Name: "Liga Portugal 2 Qualification", Country: "Portugal"},
		{ID: 64, Name: "Scottish Premiership", Country: "Scotland"},
		{ID: 135, Name: "Super League 1", Country: "Greece"},
		{ID: 46, Name: "Superligaen", Country: "Denmark"},
		{ID: 85, Name: "1. Division", Country: "Denmark"},
		{ID: 59, Name: "Eliteserien", Country: "Norway"},
		{ID: 203, Name: "1. Divisjon", Country: "Norway"},
		{ID: 71, Name: "Süper Lig", Country: "Turkey"},
		{ID: 69, Name: "Swiss Super League", Country: "Switzerland"},
		{ID: 63, Name: "Russian Premier League", Country: "Russia"},
		{ID: 441, Name: "Ukrainian Premier League", Country: "Ukraine"},
		// European Competitions
		{ID: 42, Name: "UEFA Champions League", Country: "Europe"},
		{ID: 10216, Name: "UEFA Conference League", Country: "Europe"},
		{ID: 73, Name: "UEFA Europa League", Country: "Europe"},
		{ID: 50, Name: "UEFA Euro", Country: "Europe"},
		{ID: 292, Name: "UEFA Women's Euro", Country: "Europe"},
		{ID: 9375, Name: "Women's UEFA Champions League", Country: "Europe"},
		// Domestic Cups
		{ID: 138, Name: "Copa del Rey", Country: "Spain"},
		{ID: 139, Name: "Supercopa de España", Country: "Spain"},
		{ID: 132, Name: "FA Cup", Country: "England"},
		{ID: 209, Name: "DFB Pokal", Country: "Germany"},
		{ID: 10650, Name: "Women's DFB Pokal", Country: "Germany"},
		{ID: 141, Name: "Coppa Italia", Country: "Italy"},
		{ID: 134, Name: "Coupe de France", Country: "France"},
		//Portugal Cups
		{ID: 186, Name: "Taça de Portugal", Country: "Portugal"},
		{ID: 187, Name: "Taça da Liga", Country: "Portugal"},
		{ID: 188, Name: "Supertaça Cândido de Oliveira", Country: "Portugal"},
	},
	RegionAmerica: {
		// South America
		{ID: 268, Name: "Brasileirão Série A", Country: "Brazil"},
		{ID: 8814, Name: "Brasileirão Série B", Country: "Brazil"},
		{ID: 9067, Name: "Copa do Brasil", Country: "Brazil"},
		{ID: 10077, Name: "Supercopa do Brasil", Country: "Brazil"},
		{ID: 10244, Name: "Paulista", Country: "Brazil"},
		{ID: 10272, Name: "Carioca", Country: "Brazil"},
		{ID: 10273, Name: "Mineiro", Country: "Brazil"},
		{ID: 10274, Name: "Gaúcho", Country: "Brazil"},
		{ID: 44, Name: "Copa America", Country: "South America"},
		{ID: 9490, Name: "Copa Colombia", Country: "Colombia"},
		{ID: 45, Name: "Copa Libertadores", Country: "South America"},
		{ID: 299, Name: "Copa Sudamericana", Country: "South America"},
		{ID: 491, Name: "Recopa Sudamericana", Country: "South America"},
		{ID: 112, Name: "Liga Profesional", Country: "Argentina"},
		{ID: 274, Name: "Primera A", Country: "Colombia"},
		{ID: 9125, Name: "Primera B", Country: "Colombia"},
		{ID: 161, Name: "Primera Division", Country: "Uruguay"},
		{ID: 273, Name: "Primera Division", Country: "Chile"},
		{ID: 131, Name: "Liga 1", Country: "Peru"},
		{ID: 246, Name: "Serie A", Country: "Ecuador"},
		// North America
		{ID: 130, Name: "MLS", Country: "USA"},
		{ID: 9134, Name: "NWSL", Country: "USA"},
		{ID: 230, Name: "Liga MX", Country: "Mexico"},
	},
	RegionGlobal: {
		// Middle East
		{ID: 536, Name: "Saudi Pro League", Country: "Saudi Arabia"},
		// Asia
		{ID: 9478, Name: "Indian Super League", Country: "India"},
		{ID: 223, Name: "J. League", Country: "Japan"},
		{ID: 9080, Name: "K League 1", Country: "South Korea"},
		{ID: 9137, Name: "Chinese League One", Country: "China"},
		{ID: 535, Name: "Qatar Stars League", Country: "Qatar"},
		// Oceania
		{ID: 113, Name: "A-League", Country: "Australia"},
		// Africa
		{ID: 519, Name: "Egyptian Premier League", Country: "Egypt"},
		{ID: 537, Name: "Premier Soccer League", Country: "South Africa"},
		{ID: 530, Name: "Botola Pro", Country: "Morocco"},
		// International Competitions
		{ID: 289, Name: "Africa Cup of Nations", Country: "International"},
		{ID: 77, Name: "FIFA World Cup", Country: "International"},
		{ID: 76, Name: "Women's FIFA World Cup", Country: "International"},
		{ID: 78, Name: "FIFA Club World Cup", Country: "International"},
		{ID: 9806, Name: "UEFA Nations League", Country: "International"},
		{ID: 489, Name: "Club Friendlies", Country: "International"},
		{ID: 114, Name: "International Friendlies", Country: "International"},
	},
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

// ActiveLeagueIDs returns the league IDs that should be used for API calls.
// If no leagues are selected in settings, returns the default leagues (not all).
func ActiveLeagueIDs() []int {
	settings, err := LoadSettings()
	if err != nil || len(settings.SelectedLeagues) == 0 {
		// Return default leagues for efficient API usage
		return DefaultLeagueIDs
	}

	return settings.SelectedLeagues
}

// AllLeagueIDs returns all supported league IDs (used as fallback).
func AllLeagueIDs() []int {
	totalLeagues := 0
	for _, leagues := range AllSupportedLeagues {
		totalLeagues += len(leagues)
	}

	ids := make([]int, 0, totalLeagues)
	for _, leagues := range AllSupportedLeagues {
		for _, league := range leagues {
			ids = append(ids, league.ID)
		}
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

// GetAllRegions returns a list of all available regions in order.
func GetAllRegions() []string {
	return []string{RegionEurope, RegionAmerica, RegionGlobal}
}

// GetLeaguesForRegion returns all leagues for a specific region.
func GetLeaguesForRegion(region string) []LeagueInfo {
	return AllSupportedLeagues[region]
}

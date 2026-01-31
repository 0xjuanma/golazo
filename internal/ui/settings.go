package ui

import (
	"fmt"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Settings view uses the same neon colors as the rest of the app (red/cyan theme).
// Minimal design without heavy borders.

// SettingsState holds the state for the settings view.
type SettingsState struct {
	List          list.Model        // List component for league navigation
	Selected      map[int]bool      // Map of league ID -> selected
	Leagues       []data.LeagueInfo // All leagues for current region
	AllLeagues    []data.LeagueInfo // All leagues across all regions
	Regions       []string          // Available regions
	CurrentRegion int               // Index of current region
	HasChanges    bool              // Whether there are unsaved changes
}

// NewSettingsState creates a new settings state with current saved preferences.
func NewSettingsState() *SettingsState {
	settings, _ := data.LoadSettings()

	selected := make(map[int]bool)

	// If no leagues are selected in settings, none are checked
	// User sees unchecked = will use default leagues (Premier, La Liga, UCL)
	if len(settings.SelectedLeagues) > 0 {
		for _, id := range settings.SelectedLeagues {
			selected[id] = true
		}
	}

	regions := data.GetAllRegions()
	currentRegion := 0 // Start with first region (Europe)

	// Get all leagues for current region
	leagues := data.GetLeaguesForRegion(regions[currentRegion])

	// Get all leagues across all regions for saving/loading
	allLeagueInfos := make([]data.LeagueInfo, 0, len(data.AllLeagueIDs()))
	for _, region := range regions {
		allLeagueInfos = append(allLeagueInfos, data.GetLeaguesForRegion(region)...)
	}

	// Create list items for current region
	items := make([]list.Item, len(leagues))
	for i, league := range leagues {
		items[i] = LeagueListItem{
			League:   league,
			Selected: selected[league.ID],
		}
	}

	// Create and configure the list
	delegate := NewLeagueListDelegate()
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.Filter = list.DefaultFilter
	l.SetShowHelp(false) // We use our own help text

	// Apply filter input styles
	filterCursorStyle, filterPromptStyle := FilterInputStyles()
	l.Styles.FilterCursor = filterCursorStyle
	l.FilterInput.PromptStyle = filterPromptStyle
	l.FilterInput.Cursor.Style = filterCursorStyle

	return &SettingsState{
		List:          l,
		Selected:      selected,
		Leagues:       leagues,
		AllLeagues:    allLeagueInfos,
		Regions:       regions,
		CurrentRegion: currentRegion,
	}
}

// Toggle toggles the selection state of the currently highlighted league.
func (s *SettingsState) Toggle() {
	if item, ok := s.List.SelectedItem().(LeagueListItem); ok {
		s.Selected[item.League.ID] = !s.Selected[item.League.ID]
		s.HasChanges = true
		s.refreshListItems()
	}
}

// refreshListItems updates the list items to reflect current selection state for the current region.
func (s *SettingsState) refreshListItems() {
	items := make([]list.Item, len(s.Leagues))
	for i, league := range s.Leagues {
		items[i] = LeagueListItem{
			League:   league,
			Selected: s.Selected[league.ID],
		}
	}
	s.List.SetItems(items)
}

// switchToRegion switches to a different region and updates the league list.
func (s *SettingsState) switchToRegion(regionIndex int) {
	if regionIndex < 0 || regionIndex >= len(s.Regions) {
		return
	}

	s.CurrentRegion = regionIndex
	s.Leagues = data.GetLeaguesForRegion(s.Regions[regionIndex])
	s.refreshListItems()

	// Reset filter when switching regions
	s.List.ResetFilter()
}

// NextRegion switches to the next region (with wraparound).
func (s *SettingsState) NextRegion() {
	nextRegion := (s.CurrentRegion + 1) % len(s.Regions)
	s.switchToRegion(nextRegion)
}

// PreviousRegion switches to the previous region (with wraparound).
func (s *SettingsState) PreviousRegion() {
	prevRegion := s.CurrentRegion - 1
	if prevRegion < 0 {
		prevRegion = len(s.Regions) - 1
	}
	s.switchToRegion(prevRegion)
}

// Save persists the current selection to settings.yaml.
func (s *SettingsState) Save() error {
	var selectedIDs []int
	for _, league := range s.AllLeagues {
		if s.Selected[league.ID] {
			selectedIDs = append(selectedIDs, league.ID)
		}
	}

	settings := &data.Settings{
		SelectedLeagues: selectedIDs,
	}

	err := data.SaveSettings(settings)
	if err == nil {
		s.HasChanges = false
	}
	return err
}

// SelectedCount returns the number of selected leagues.
func (s *SettingsState) SelectedCount() int {
	count := 0
	for _, isSelected := range s.Selected {
		if isSelected {
			count++
		}
	}
	return count
}

// Fixed width for settings panel (30% wider than original 48)
const settingsBoxWidth = 62

// renderTabBar renders the regional tabs at the top of the settings view.
func renderTabBar(regions []string, currentRegion int, width int) string {
	var tabElements []string

	for i, region := range regions {
		var tabStyle lipgloss.Style

		if i == currentRegion {
			// Active tab - neon cyan
			tabStyle = lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true).
				Padding(0, 2)
		} else {
			// Inactive tab - dim
			tabStyle = lipgloss.NewStyle().
				Foreground(neonDim).
				Padding(0, 2)
		}

		tabElements = append(tabElements, tabStyle.Render(region))
	}

	// Join tabs with separator
	tabs := lipgloss.JoinHorizontal(lipgloss.Left, tabElements...)

	// Center the tab bar
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(tabs)
}

// RenderSettingsView renders the settings view for league customization.
// Uses minimal styling consistent with the rest of the app (red/cyan neon theme).
// bannerType determines what status banner (if any) to display at the top.
func RenderSettingsView(width, height int, state *SettingsState, bannerType constants.StatusBannerType) string {
	if state == nil {
		return ""
	}

	// Calculate available space for the list
	const (
		titleHeight  = 3 // Title + margin
		tabsHeight   = 2 // Tab bar + margin
		infoHeight   = 2 // Selection info
		helpHeight   = 2 // Help text
		extraPadding = 4 // Additional vertical spacing
	)

	listWidth := settingsBoxWidth
	listHeight := height - titleHeight - tabsHeight - infoHeight - helpHeight - extraPadding
	if listHeight < 5 {
		listHeight = 5
	}

	// Update list dimensions
	state.List.SetSize(listWidth, listHeight)

	// Add status banner if needed
	statusBanner := renderStatusBanner(bannerType, settingsBoxWidth)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	// Title - compact header with gradient and diagonal fill
	title := design.RenderHeader(constants.PanelLeaguePreferences, settingsBoxWidth)

	// Render the tab bar
	tabs := renderTabBar(state.Regions, state.CurrentRegion, settingsBoxWidth)

	// Render the list
	listContent := state.List.View()
	listContainerStyle := lipgloss.NewStyle().Width(settingsBoxWidth)
	listContent = listContainerStyle.Render(listContent)

	// Selection info
	selectedCount := state.SelectedCount()
	var infoText string
	if selectedCount == 0 {
		infoText = "No selection = default leagues"
	} else {
		infoText = fmt.Sprintf("%d of %d selected", selectedCount, len(state.AllLeagues))
	}
	infoStyle := neonDimStyle.Width(settingsBoxWidth).Align(lipgloss.Center)
	info := infoStyle.Render(infoText)

	// Help text - update to include tab navigation
	helpText := constants.HelpSettingsView
	helpStyle := neonDimStyle.Width(settingsBoxWidth).Align(lipgloss.Center)
	help := helpStyle.Render(helpText)

	// Combine content (minimal, no borders)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		statusBanner,
		title,
		"",
		tabs,
		"",
		listContent,
		"",
		info,
		help,
	)

	// Center in the terminal
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

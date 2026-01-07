package ui

import (
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Use consolidated neon colors from neon_styles.go
// These aliases are kept for backward compatibility but reference the main color definitions
var (
	delegateNeonRed   = neonRed
	delegateNeonCyan  = neonCyan
	delegateNeonWhite = neonWhite
	delegateNeonGray  = neonDim
	delegateNeonDim   = neonDimGray
)

// NewMatchListDelegate creates a custom list delegate for match items.
// Height is set to 3 to accommodate title + 2-line description (with KO time).
// Uses Neon Gradient styling: red title, cyan description on selection.
func NewMatchListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Set height to 3 lines: title (1) + description with KO time (2)
	d.SetHeight(3)

	// Use consolidated neon colors from neon_styles.go

	// Selected items: Neon red title, cyan description, red left border
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(neonRed).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(neonRed)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(neonCyan).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(neonRed)

	// Normal items: White title, gray description
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(neonWhite).
		Padding(0, 1)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(neonGray).
		Padding(0, 1)

	// Dimmed items (non-matching during filter): very dim
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)
	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)

	// Filter match highlight: cyan bold for matched text
	d.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true).
		Underline(true)

	return d
}

// LeagueListDelegate is a custom delegate that renders checkboxes separately from titles.
// This fixes the filter cursor positioning issue by keeping the checkbox out of the title.
type LeagueListDelegate struct {
	list.DefaultDelegate
}

// Render renders a league list item with a checkbox prefix.
// The checkbox is rendered separately from the title to prevent filter cursor shift.
func (d LeagueListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	leagueItem, ok := item.(LeagueListItem)
	if !ok {
		// Fallback: render without checkbox if not a LeagueListItem
		// This shouldn't happen in normal usage, but handle gracefully
		title := item.FilterValue()
		desc := ""
		if descItem, ok := item.(interface{ Description() string }); ok {
			desc = descItem.Description()
		}

		var titleStyle, descStyle lipgloss.Style
		if index == m.Index() {
			titleStyle = d.Styles.SelectedTitle
			descStyle = d.Styles.SelectedDesc
		} else {
			titleStyle = d.Styles.NormalTitle
			descStyle = d.Styles.NormalDesc
		}

		titleRendered := titleStyle.Render(title)
		descRendered := descStyle.Render(desc)
		result := lipgloss.JoinVertical(lipgloss.Left, titleRendered, descRendered)
		_, _ = w.Write([]byte(result))
		return
	}

	// Get checkbox state
	checkbox := "[ ]"
	if leagueItem.Selected {
		checkbox = "[x]"
	}

	// Check if item matches filter by comparing filter value with item's FilterValue
	filterValue := m.FilterValue()
	isFiltering := m.FilterState() == list.Filtering
	isDimmed := isFiltering && filterValue != "" && !d.itemMatchesFilter(leagueItem, filterValue)

	// Render checkbox with appropriate styling based on selection and filter state
	var checkboxStyle lipgloss.Style
	if index == m.Index() {
		// Selected item - checkbox uses selected title style color
		checkboxStyle = lipgloss.NewStyle().
			Foreground(delegateNeonRed).
			Bold(true)
	} else if isDimmed {
		// Dimmed item - checkbox uses dimmed style color
		checkboxStyle = lipgloss.NewStyle().
			Foreground(delegateNeonDim)
	} else {
		// Normal item - checkbox uses normal title style color
		checkboxStyle = lipgloss.NewStyle().
			Foreground(delegateNeonWhite)
	}

	// Render checkbox
	checkboxRendered := checkboxStyle.Render(checkbox + " ")

	// Get the title and description from the item
	title := leagueItem.Title()
	desc := leagueItem.Description()

	// Apply appropriate styles based on selection and filter state
	var titleStyle, descStyle lipgloss.Style
	if index == m.Index() {
		// Selected item
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc
	} else if isDimmed {
		// Dimmed item (filtered out)
		titleStyle = d.Styles.DimmedTitle
		descStyle = d.Styles.DimmedDesc
	} else {
		// Normal item
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc
	}

	// Apply filter match highlighting to title if filtering and item matches
	if isFiltering && !isDimmed && filterValue != "" {
		title = d.HighlightMatches(title, filterValue)
	}

	// Render title and description
	titleRendered := titleStyle.Render(title)
	descRendered := descStyle.Render(desc)

	// Combine checkbox + title on first line, description on second line
	titleLine := lipgloss.JoinHorizontal(lipgloss.Left, checkboxRendered, titleRendered)
	result := lipgloss.JoinVertical(lipgloss.Left, titleLine, descRendered)
	_, _ = w.Write([]byte(result))
}

// itemMatchesFilter checks if an item matches the filter value.
func (d LeagueListDelegate) itemMatchesFilter(item LeagueListItem, filterValue string) bool {
	if filterValue == "" {
		return true
	}
	filterLower := strings.ToLower(filterValue)
	itemFilterValue := strings.ToLower(item.FilterValue())
	return strings.Contains(itemFilterValue, filterLower)
}

// HighlightMatches highlights matching text in the title using FilterMatch style.
func (d LeagueListDelegate) HighlightMatches(text, filterValue string) string {
	if filterValue == "" {
		return text
	}

	// Simple case-insensitive matching
	lowerText := strings.ToLower(text)
	lowerFilter := strings.ToLower(filterValue)

	if !strings.Contains(lowerText, lowerFilter) {
		return text
	}

	// Find all matches and highlight them
	var result strings.Builder
	lastIndex := 0
	textRunes := []rune(text)
	filterRunes := []rune(filterValue)

	for {
		// Find next match
		matchIndex := -1
		for i := lastIndex; i <= len(textRunes)-len(filterRunes); i++ {
			match := true
			for j := 0; j < len(filterRunes); j++ {
				if !strings.EqualFold(string(textRunes[i+j]), string(filterRunes[j])) {
					match = false
					break
				}
			}
			if match {
				matchIndex = i
				break
			}
		}

		if matchIndex == -1 {
			// No more matches, append rest of text
			result.WriteString(string(textRunes[lastIndex:]))
			break
		}

		// Append text before match
		result.WriteString(string(textRunes[lastIndex:matchIndex]))

		// Append highlighted match
		matchText := string(textRunes[matchIndex : matchIndex+len(filterRunes)])
		highlighted := d.Styles.FilterMatch.Render(matchText)
		result.WriteString(highlighted)

		lastIndex = matchIndex + len(filterRunes)
	}

	return result.String()
}

// NewLeagueListDelegate creates a custom list delegate for league selection.
// Height is set to 2 to show league name (with checkbox) and country.
// Uses same red/cyan neon styling as match delegate for consistency.
// The checkbox is rendered separately from the title to fix filter cursor positioning.
func NewLeagueListDelegate() LeagueListDelegate {
	d := LeagueListDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}

	// Set height to 2 lines: title with checkbox (1) + country (1)
	d.SetHeight(2)

	// Selected items: Neon red title, cyan description, red left border
	// Matches the match list delegate exactly
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(delegateNeonRed).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(delegateNeonRed)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(delegateNeonCyan).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(delegateNeonRed)

	// Normal items: White title, gray description
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(delegateNeonWhite).
		Padding(0, 1)
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(delegateNeonGray).
		Padding(0, 1)

	// Dimmed items (non-matching during filter): very dim
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(delegateNeonDim).
		Padding(0, 1)
	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(delegateNeonDim).
		Padding(0, 1)

	// Filter match highlight: cyan bold for matched text
	d.Styles.FilterMatch = lipgloss.NewStyle().
		Foreground(delegateNeonCyan).
		Bold(true).
		Underline(true)

	return d
}

package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const statisticsDialogID = "statistics"

// StatisticsDialog displays all match statistics in a comparison view.
type StatisticsDialog struct {
	homeTeam    string
	awayTeam    string
	homeScore	int
	awayScore	int
	statistics  []api.MatchStatistic
	scrollIndex int
	maxVisible  int
}

// NewStatisticsDialog creates a new statistics dialog.
func NewStatisticsDialog(homeTeam, awayTeam string, homeScore, awayScore int, statistics []api.MatchStatistic) *StatisticsDialog {
	return &StatisticsDialog{
		homeTeam:    homeTeam,
		awayTeam:    awayTeam,
		homeScore:    homeScore,
		awayScore:    awayScore,
		statistics:  statistics,
		scrollIndex: 0,
		maxVisible:  20, // Number of stats visible at once (larger dialog)
	}
}

// ID returns the dialog identifier.
func (d *StatisticsDialog) ID() string {
	return statisticsDialogID
}

// Update handles input for the statistics dialog.
func (d *StatisticsDialog) Update(msg tea.Msg) (Dialog, DialogAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "x", "q":
			return d, DialogActionClose{}
		case "j", "down":
			maxScroll := len(d.statistics) - d.maxVisible
			maxScroll = max(maxScroll, 0)
			if d.scrollIndex < maxScroll {
				d.scrollIndex++
			}
		case "k", "up":
			if d.scrollIndex > 0 {
				d.scrollIndex--
			}
		}
	}
	return d, nil
}

// View renders the statistics comparison.
func (d *StatisticsDialog) View(width, height int) string {
	// Larger dimensions for better readability
	dialogWidth, dialogHeight := DialogSize(width, height, 97, 36)

	// Build the content
	content := d.renderContent(dialogWidth - 6) // Account for padding and border

	return RenderDialogFrameWithHelp(constants.PanelMatchStatistics, content, constants.HelpStatisticsDialog, dialogWidth, dialogHeight)
}

// renderContent renders the statistics content.
func (d *StatisticsDialog) renderContent(width int) string {
	if len(d.statistics) == 0 {
		return dialogDimStyle.Render("No statistics available")
	}

	var lines []string

	// Team header
	header := d.renderTeamHeader(width)
	lines = append(lines, header)
	lines = append(lines, "")

	// Separator
	separator := dialogSeparatorStyle.Render(strings.Repeat("─", width))
	lines = append(lines, separator)

	// Calculate visible range
	endIdx := d.scrollIndex + d.maxVisible
	endIdx = min(endIdx, len(d.statistics))

	// Render visible statistics
	for i := d.scrollIndex; i < endIdx; i++ {
		stat := d.statistics[i]
		statLine := d.renderStatRow(stat, width)
		lines = append(lines, statLine)
	}

	// Scroll indicator if needed
	if len(d.statistics) > d.maxVisible {
		scrollInfo := fmt.Sprintf("(%d-%d of %d)", d.scrollIndex+1, endIdx, len(d.statistics))
		lines = append(lines, "")
		lines = append(lines, dialogDimStyle.Render(scrollInfo))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderTeamHeader renders the team names header.
func (d *StatisticsDialog) renderTeamHeader(width int) string {
	// Truncate team names if needed
	homeTeam := d.homeTeam
	awayTeam := d.awayTeam
	maxLen := (width - 10) / 2
	if len(homeTeam) > maxLen {
		homeTeam = homeTeam[:maxLen-1] + "…"
	}
	if len(awayTeam) > maxLen {
		awayTeam = awayTeam[:maxLen-1] + "…"
	}

	headerText := fmt.Sprintf("%s %d - %d  %s", homeTeam, d.homeScore , d.awayScore, awayTeam)
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(neonCyan).
		Bold(true).
		Render(headerText)
}

// renderStatRow renders a single statistic row with comparison bar.
func (d *StatisticsDialog) renderStatRow(stat api.MatchStatistic, width int) string {
	// Parse values for comparison
	homeVal := parseStatNumber(stat.HomeValue)
	awayVal := parseStatNumber(stat.AwayValue)

	// Format label
	label := stat.Label
	if label == "" {
		label = stat.Key
	}
	maxLabelLen := 20
	if len(label) > maxLabelLen {
		label = label[:maxLabelLen-1] + "…"
	}

	// Fixed width for values to ensure alignment
	valWidth := 12

	// Truncate long values if needed
	homeValStr := stat.HomeValue
	awayValStr := stat.AwayValue
	if len(homeValStr) > valWidth {
		homeValStr = homeValStr[:valWidth-1] + "…"
	}
	if len(awayValStr) > valWidth {
		awayValStr = awayValStr[:valWidth-1] + "…"
	}

	// Calculate bar widths
	barWidth := 16
	homeBarWidth, awayBarWidth := calculateBarWidths(homeVal, awayVal, barWidth)

	barColor := neonCyan
	if homeVal == awayVal {
		barColor = neonWhite
	}
	// Render solid color bars (cyan for home, gray for away)
	homeBar :=  lipgloss.NewStyle().Foreground(neonGray).Render(strings.Repeat("░", barWidth-homeBarWidth)) +
				lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", homeBarWidth))

	awayBar := 	lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", awayBarWidth)) +
				lipgloss.NewStyle().Foreground(neonGray).Render(strings.Repeat("░", barWidth-awayBarWidth))

	homeBarStyled := lipgloss.NewStyle().Foreground(neonCyan).Render(homeBar)
	awayBarStyled := lipgloss.NewStyle().Foreground(neonGray).Render(awayBar)

	// Style values - bold cyan for winner, dim for loser
	winnerStyle := lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
	loserStyle := dialogDimStyle

	var homeStyled, awayStyled string
	if homeVal > awayVal {
		homeStyled = winnerStyle.Width(valWidth).Align(lipgloss.Right).Render(homeValStr)
		awayStyled = loserStyle.Width(valWidth).Align(lipgloss.Left).Render(awayValStr)
	} else if awayVal > homeVal {
		homeStyled = loserStyle.Width(valWidth).Align(lipgloss.Right).Render(homeValStr)
		awayStyled = winnerStyle.Width(valWidth).Align(lipgloss.Left).Render(awayValStr)
	} else {
		// Tie - both normal
		homeStyled = dialogValueStyle.Width(valWidth).Align(lipgloss.Right).Render(homeValStr)
		awayStyled = dialogValueStyle.Width(valWidth).Align(lipgloss.Left).Render(awayValStr)
	}

	// Build the row with fixed widths
	labelStyled := dialogLabelStyle.Width(maxLabelLen).Render(label)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyled,
		" ",
		homeStyled,
		" ",
		homeBarStyled,
		"│",
		awayBarStyled,
		" ",
		awayStyled,
	)
}

// calculateBarWidths calculates proportional bar widths for two values.
func calculateBarWidths(home, away float64, maxWidth int) (int, int) {
	total := home + away
	if total == 0 {
		return maxWidth / 2, maxWidth / 2
	}

	homeWidth := int((home / total) * float64(maxWidth))
	awayWidth := int((away / total) * float64(maxWidth))

	// Ensure at least 1 bar segment if value > 0
	if home > 0 && homeWidth == 0 {
		homeWidth = 1
	}
	if away > 0 && awayWidth == 0 {
		awayWidth = 1
	}

	// Cap at maxWidth
	if homeWidth > maxWidth {
		homeWidth = maxWidth
	}
	if awayWidth > maxWidth {
		awayWidth = maxWidth
	}

	return homeWidth, awayWidth
}

// parseStatNumber extracts a numeric value from a stat string.
func parseStatNumber(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")

	// Handle formats like "23 (45%)" - take first number
	if idx := strings.Index(s, " "); idx > 0 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "("); idx > 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const standingsDialogID = "standings"

// StandingsDialog displays the league standings table for a match.
type StandingsDialog struct {
	leagueName  string
	standings   []api.LeagueTableEntry
	homeTeamID  int
	awayTeamID  int
	scrollIndex int
}

// NewStandingsDialog creates a new standings dialog.
func NewStandingsDialog(leagueName string, standings []api.LeagueTableEntry, homeTeamID, awayTeamID int) *StandingsDialog {
	return &StandingsDialog{
		leagueName:  leagueName,
		standings:   standings,
		homeTeamID:  homeTeamID,
		awayTeamID:  awayTeamID,
		scrollIndex: 0,
	}
}

// ID returns the dialog identifier.
func (d *StandingsDialog) ID() string {
	return standingsDialogID
}

// Update handles input for the standings dialog.
func (d *StandingsDialog) Update(msg tea.Msg) (Dialog, DialogAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "s", "q":
			return d, DialogActionClose{}
		case "j", "down":
			if d.scrollIndex < len(d.standings)-1 {
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

// View renders the standings table.
func (d *StandingsDialog) View(width, height int) string {
	// Calculate dialog dimensions (larger for better readability)
	dialogWidth, dialogHeight := DialogSize(width, height, 90, 32)

	// Build the table content
	content := d.renderTable(dialogWidth - 6) // Account for padding and border

	return RenderDialogFrameWithHelp(d.leagueName+" Standings", content, constants.HelpStandingsDialog, dialogWidth, dialogHeight)
}

// renderTable renders the standings table.
func (d *StandingsDialog) renderTable(width int) string {
	if len(d.standings) == 0 {
		return dialogDimStyle.Render("No standings data available")
	}

	var lines []string

	// Header row
	header := d.renderHeaderRow(width)
	lines = append(lines, header)

	// Separator
	separator := dialogSeparatorStyle.Render(strings.Repeat("─", width))
	lines = append(lines, separator)

	// Data rows
	for _, entry := range d.standings {
		row := d.renderTeamRow(entry, width)
		lines = append(lines, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Column widths for consistent alignment
const (
	standingsColPos  = 4 // Position column
	standingsColStat = 5 // Stat columns (P, W, D, L)
	standingsColGD   = 5 // Goal difference (needs +/- sign)
	standingsColPts  = 5 // Points column
)

// renderHeaderRow renders the table header.
func (d *StandingsDialog) renderHeaderRow(width int) string {
	teamWidth := width - standingsColPos - (standingsColStat * 4) - standingsColGD - standingsColPts - 4

	return lipgloss.JoinHorizontal(lipgloss.Top,
		dialogHeaderStyle.Width(standingsColPos).Align(lipgloss.Right).Render("#"),
		"  ",
		dialogHeaderStyle.Width(teamWidth).Align(lipgloss.Left).Render("Team"),
		dialogHeaderStyle.Width(standingsColStat).Align(lipgloss.Right).Render("P"),
		dialogHeaderStyle.Width(standingsColStat).Align(lipgloss.Right).Render("W"),
		dialogHeaderStyle.Width(standingsColStat).Align(lipgloss.Right).Render("D"),
		dialogHeaderStyle.Width(standingsColStat).Align(lipgloss.Right).Render("L"),
		dialogHeaderStyle.Width(standingsColGD).Align(lipgloss.Right).Render("GD"),
		dialogHeaderStyle.Width(standingsColPts).Align(lipgloss.Right).Render("Pts"),
	)
}

// renderTeamRow renders a single team row.
func (d *StandingsDialog) renderTeamRow(entry api.LeagueTableEntry, width int) string {
	isHighlighted := entry.Team.ID == d.homeTeamID || entry.Team.ID == d.awayTeamID

	teamWidth := width - standingsColPos - (standingsColStat * 4) - standingsColGD - standingsColPts - 4

	// Truncate team name if needed
	teamName := entry.Team.ShortName
	if teamName == "" {
		teamName = entry.Team.Name
	}
	if len(teamName) > teamWidth-1 {
		teamName = teamName[:teamWidth-2] + "…"
	}

	// Format goal difference with sign
	gdStr := formatGoalDifference(entry.GoalDifference)

	// Build row content with fixed widths
	rowContent := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(standingsColPos).Align(lipgloss.Right).Render(fmt.Sprintf("%d", entry.Position)),
		"  ",
		lipgloss.NewStyle().Width(teamWidth).Align(lipgloss.Left).Render(teamName),
		lipgloss.NewStyle().Width(standingsColStat).Align(lipgloss.Right).Render(fmt.Sprintf("%d", entry.Played)),
		lipgloss.NewStyle().Width(standingsColStat).Align(lipgloss.Right).Render(fmt.Sprintf("%d", entry.Won)),
		lipgloss.NewStyle().Width(standingsColStat).Align(lipgloss.Right).Render(fmt.Sprintf("%d", entry.Drawn)),
		lipgloss.NewStyle().Width(standingsColStat).Align(lipgloss.Right).Render(fmt.Sprintf("%d", entry.Lost)),
		lipgloss.NewStyle().Width(standingsColGD).Align(lipgloss.Right).Render(gdStr),
		lipgloss.NewStyle().Width(standingsColPts).Align(lipgloss.Right).Render(fmt.Sprintf("%d", entry.Points)),
	)

	// Apply row styling
	if isHighlighted {
		// Background highlight for match teams
		return lipgloss.NewStyle().
			Background(neonDark).
			Foreground(neonCyan).
			Bold(true).
			Width(width).
			Render(rowContent)
	}

	return dialogValueStyle.Render(rowContent)
}

// formatGoalDifference formats goal difference with +/- sign.
func formatGoalDifference(gd int) string {
	if gd > 0 {
		return fmt.Sprintf("+%d", gd)
	}
	return fmt.Sprintf("%d", gd)
}

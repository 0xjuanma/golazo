package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TopScorersDialog displays the World Cup top scorers table.
type TopScorersDialog struct {
	scorers     []api.WCTopScorer
	scrollIndex int
}

// NewTopScorersDialog creates a new top scorers dialog.
func NewTopScorersDialog(scorers []api.WCTopScorer) *TopScorersDialog {
	return &TopScorersDialog{
		scorers:     scorers,
		scrollIndex: 0,
	}
}

// ID returns the dialog identifier.
func (d *TopScorersDialog) ID() string {
	return TopScorersDialogID
}

// Update handles input for the top scorers dialog.
func (d *TopScorersDialog) Update(msg tea.Msg) (Dialog, DialogAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "s", "q":
			return d, DialogActionClose{}
		case "j", "down":
			d.scrollIndex = scrollDown(d.scrollIndex, len(d.scorers)-1)
		case "k", "up":
			d.scrollIndex = scrollUp(d.scrollIndex)
		}
	}
	return d, nil
}

// View renders the top scorers table.
func (d *TopScorersDialog) View(width, height int) string {
	dialogWidth, dialogHeight := DialogSize(width, height, 72, 34)
	innerWidth := dialogWidth - 6 // account for padding and border
	content := d.renderTable(innerWidth, dialogHeight-8)
	return RenderDialogFrameWithHelp("World Cup Top Scorers", content, constants.HelpTopScorersDialog, dialogWidth, dialogHeight)
}

// Column widths
const (
	scorerColRank  = 4
	scorerColGoals = 6
)

func (d *TopScorersDialog) renderTable(width, visibleRows int) string {
	if len(d.scorers) == 0 {
		return dialogDimStyle.Render("No scorer data available")
	}

	if visibleRows < 1 {
		visibleRows = 10
	}

	teamWidth := 16
	nameWidth := width - scorerColRank - teamWidth - scorerColGoals - 4

	var lines []string

	// Header
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		dialogHeaderStyle.Width(scorerColRank).Align(lipgloss.Right).Render("#"),
		"  ",
		dialogHeaderStyle.Width(nameWidth).Align(lipgloss.Left).Render("Player"),
		dialogHeaderStyle.Width(teamWidth).Align(lipgloss.Left).Render("Team"),
		dialogHeaderStyle.Width(scorerColGoals).Align(lipgloss.Right).Render("Goals"),
	)
	lines = append(lines, header)

	separator := dialogSeparatorStyle.Render(strings.Repeat("─", width))
	lines = append(lines, separator)

	// Visible window
	start := d.scrollIndex
	end := start + visibleRows
	if end > len(d.scorers) {
		end = len(d.scorers)
	}

	for i, s := range d.scorers[start:end] {
		rank := start + i + 1
		lines = append(lines, d.renderScorerRow(rank, s, nameWidth, teamWidth, width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (d *TopScorersDialog) renderScorerRow(rank int, s api.WCTopScorer, nameWidth, teamWidth, rowWidth int) string {
	name := truncateString(s.PlayerName, nameWidth-1)
	team := truncateString(s.Team, teamWidth-1)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		dialogAlignRight(scorerColRank, fmt.Sprintf("%d", rank)),
		"  ",
		dialogAlignLeft(nameWidth, name),
		dialogAlignLeft(teamWidth, team),
		dialogAlignRight(scorerColGoals, fmt.Sprintf("%d", s.Goals)),
	)

	if rank == 1 {
		return lipgloss.NewStyle().
			Background(neonDark).
			Foreground(neonCyan).
			Bold(true).
			Width(rowWidth).
			Render(row)
	}
	return dialogValueStyle.Render(row)
}

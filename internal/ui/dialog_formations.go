package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const formationsDialogID = "formations"

// FormationsDialog displays the match formations for both teams.
type FormationsDialog struct {
	homeTeam      string
	awayTeam      string
	homeFormation string
	awayFormation string
	homeStarting  []api.PlayerInfo
	awayStarting  []api.PlayerInfo
	focusedTeam   int // 0 = home, 1 = away
}

// NewFormationsDialog creates a new formations dialog.
func NewFormationsDialog(
	homeTeam, awayTeam string,
	homeFormation, awayFormation string,
	homeStarting, awayStarting []api.PlayerInfo,
) *FormationsDialog {
	return &FormationsDialog{
		homeTeam:      homeTeam,
		awayTeam:      awayTeam,
		homeFormation: homeFormation,
		awayFormation: awayFormation,
		homeStarting:  homeStarting,
		awayStarting:  awayStarting,
		focusedTeam:   0,
	}
}

// ID returns the dialog identifier.
func (d *FormationsDialog) ID() string {
	return formationsDialogID
}

// Update handles input for the formations dialog.
func (d *FormationsDialog) Update(msg tea.Msg) (Dialog, DialogAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "f", "q":
			return d, DialogActionClose{}
		case "tab", "h", "l", "left", "right":
			// Toggle between home and away
			d.focusedTeam = 1 - d.focusedTeam
		}
	}
	return d, nil
}

// View renders the formations view.
func (d *FormationsDialog) View(width, height int) string {
	// Larger dimensions for better readability
	dialogWidth, dialogHeight := DialogSize(width, height, 97, 36)

	// Build the content
	content := d.renderFormations(dialogWidth - 6)

	title := d.buildTitle()
	help := "tab/h/l: switch team | esc: close"
	return RenderDialogFrameWithHelp(title, content, help, dialogWidth, dialogHeight)
}

// buildTitle builds the dialog title.
func (d *FormationsDialog) buildTitle() string {
	return "Formations"
}

// renderFormations renders both team formations side by side.
func (d *FormationsDialog) renderFormations(width int) string {
	halfWidth := (width - 3) / 2 // Account for separator

	// Render each team panel
	homePanel := d.renderTeamPanel(d.homeTeam, d.homeFormation, d.homeStarting, halfWidth, d.focusedTeam == 0)
	awayPanel := d.renderTeamPanel(d.awayTeam, d.awayFormation, d.awayStarting, halfWidth, d.focusedTeam == 1)

	// Separator
	separator := dialogSeparatorStyle.Render(" │ ")

	return lipgloss.JoinHorizontal(lipgloss.Top, homePanel, separator, awayPanel)
}

// renderTeamPanel renders a single team's formation panel.
func (d *FormationsDialog) renderTeamPanel(teamName, formation string, players []api.PlayerInfo, width int, focused bool) string {
	var lines []string

	// Team header
	var headerStyle lipgloss.Style
	if focused {
		headerStyle = dialogTeamStyle.Width(width).Align(lipgloss.Center)
	} else {
		headerStyle = dialogDimStyle.Width(width).Align(lipgloss.Center)
	}

	// Truncate team name if needed
	if len(teamName) > width-2 {
		teamName = teamName[:width-3] + "…"
	}

	header := headerStyle.Render(teamName)
	lines = append(lines, header)

	// Formation string
	formationStr := formation
	if formationStr == "" {
		formationStr = "Formation N/A"
	}
	formationLine := dialogDimStyle.Width(width).Align(lipgloss.Center).Render(formationStr)
	lines = append(lines, formationLine)

	// Separator
	sep := dialogSeparatorStyle.Render(strings.Repeat("─", width))
	lines = append(lines, sep)

	// Player list
	if len(players) == 0 {
		noData := dialogDimStyle.Width(width).Align(lipgloss.Center).Render("Lineup not available")
		lines = append(lines, noData)
	} else {
		for _, player := range players {
			playerLine := d.renderPlayerLine(player, width, focused)
			lines = append(lines, playerLine)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderPlayerLine renders a single player line with number, position, and rating.
func (d *FormationsDialog) renderPlayerLine(player api.PlayerInfo, width int, focused bool) string {
	// Number
	numStr := ""
	if player.Number > 0 {
		numStr = fmt.Sprintf("%2d", player.Number)
	} else {
		numStr = "  "
	}

	// Position (abbreviated)
	posStr := player.Position
	if len(posStr) > 3 {
		posStr = posStr[:3]
	}
	posStr = fmt.Sprintf("%-3s", posStr)

	// Player name (truncated if needed)
	nameWidth := width - 14 // Account for number, position, rating badge, spacing
	name := player.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-1] + "…"
	}
	name = fmt.Sprintf("%-*s", nameWidth, name)

	// Apply styles
	var numStyle, posStyle, nameStyle lipgloss.Style
	if focused {
		numStyle = dialogValueStyle
		posStyle = dialogDimStyle
		nameStyle = dialogContentStyle
	} else {
		numStyle = dialogDimStyle
		posStyle = dialogDimStyle
		nameStyle = dialogDimStyle
	}

	// Render rating with badge for high ratings
	ratingRendered := d.renderRating(player.Rating, focused)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		numStyle.Render(numStr),
		" ",
		posStyle.Render(posStr),
		" ",
		nameStyle.Render(name),
		ratingRendered,
	)
}

// renderRating renders the player rating with color styling.
func (d *FormationsDialog) renderRating(rating string, focused bool) string {
	if rating == "" {
		return "    "
	}

	ratingStr := fmt.Sprintf("%4s", rating)

	if !focused {
		return dialogDimStyle.Render(ratingStr)
	}

	// Parse rating to determine color
	var val float64
	fmt.Sscanf(rating, "%f", &val)

	if val >= 8.0 {
		// Exceptional rating - bold red
		return lipgloss.NewStyle().Foreground(neonRed).Bold(true).Render(ratingStr)
	} else if val >= 7.5 {
		// High rating - bold cyan
		return lipgloss.NewStyle().Foreground(neonCyan).Bold(true).Render(ratingStr)
	} else if val >= 6.5 {
		// Good rating - normal cyan
		return lipgloss.NewStyle().Foreground(neonCyan).Render(ratingStr)
	} else if val >= 6.0 {
		// Average rating - white
		return dialogValueStyle.Render(ratingStr)
	}
	// Below average - dim
	return dialogDimStyle.Render(ratingStr)
}

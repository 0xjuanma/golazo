package worldcup

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderBracket renders the knockout bracket with box-drawing connectors
// between paired matchups to visually convey the bracket progression.
func RenderBracket(width, height int, wcData *api.WorldCupData, scrollOffset int, statusBanner string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if wcData == nil {
		return LoadingStyle.Render("No bracket data")
	}

	header := design.RenderHeader(wcData.Name+" — Knockout Bracket", width-2)
	help := HelpStyle.Width(width).Render("j/k: scroll  u: upcoming  s: top scorers  Esc: back to grid  q: quit")

	var lines []string

	for _, round := range wcData.KnockoutRounds {
		roundHdr := RoundHeaderStyle.Render(
			"── " + strings.ToUpper(round.Label) +
				" " + strings.Repeat("─", max(0, width-len(round.Label)-8)),
		)
		lines = append(lines, roundHdr, "")
		lines = append(lines, renderBracketRound(round, width)...)
		lines = append(lines, "")
	}

	// Bronze final
	if wcData.BronzeFinal != nil {
		bronzeHdr := RoundHeaderStyle.Render("── 3RD PLACE " + strings.Repeat("─", max(0, width-17)))
		lines = append(lines, bronzeHdr, "")
		lines = append(lines, renderBracketLine(*wcData.BronzeFinal))
		lines = append(lines, "")
	}

	// Champion card
	if wcData.Champion != nil {
		lines = append(lines, "")
		champ := ChampionStyle.Render(fmt.Sprintf("  🏆  Champion: %s", TeamLabel(*wcData.Champion)))
		lines = append(lines, champ)
	}

	// Scroll window
	overhead := 7
	if statusBanner == "" {
		overhead = 6
	}
	availableLines := height - overhead
	if availableLines < 1 {
		availableLines = 1
	}

	maxScroll := len(lines) - availableLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollOffset > maxScroll {
		scrollOffset = maxScroll
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	end := scrollOffset + availableLines
	if end > len(lines) {
		end = len(lines)
	}
	visible := strings.Join(lines[scrollOffset:end], "\n")

	scrollIndicator := ""
	if len(lines) > availableLines {
		scrollIndicator = lipgloss.NewStyle().Foreground(colorDim).
			Render(fmt.Sprintf("  (%d/%d lines)", scrollOffset+1, len(lines)))
	}

	parts := []string{}
	if statusBanner != "" {
		parts = append(parts, statusBanner)
	}
	parts = append(parts, header, "", PanelStyle.Width(width-2).Render(visible), scrollIndicator, help)

	return padToHeight(lipgloss.JoinVertical(lipgloss.Left, parts...), height)
}

// renderBracketRound renders all matchups in a round, pairing consecutive
// matchups with box-drawing connectors to show next-round opponents.
func renderBracketRound(round api.WCKnockoutRound, width int) []string {
	var lines []string
	mus := round.Matchups

	for i := 0; i < len(mus); i += 2 {
		mu1 := mus[i]
		mu1Line := renderBracketLine(mu1)
		lines = append(lines, mu1Line)

		if i+1 < len(mus) {
			mu2 := mus[i+1]
			mu2Line := renderBracketLineRaw(mu2, false)

			// Connector: ──╮ / ├─► next / ──╯
			winnerLabel := ""
			if mu1.WinnerID != nil && mu2.WinnerID != nil {
				w1 := nextRoundTeamName(mu1)
				w2 := nextRoundTeamName(mu2)
				winnerLabel = fmt.Sprintf(" ► %s vs %s", w1, w2)
			}
			connector := ConnectorStyle.Render("──╮")
			middle := ConnectorStyle.Render("├─") + lipgloss.NewStyle().Foreground(colorDim).Render(winnerLabel)
			bottom := ConnectorStyle.Render("──╯")

			// Align connectors to the column immediately after the rendered
			// match line, using lipgloss.Width to account for emoji/RIS pairs
			// whose visual width differs from byte length.
			pad := "  "
			gap := "  "
			alignCol := lipgloss.Width(pad + mu2Line + gap)
			alignPad := strings.Repeat(" ", alignCol)

			lines = append(lines,
				pad+mu2Line+gap+connector,
				alignPad+middle,
				alignPad+bottom,
			)
		}
	}
	return lines
}

// nextRoundTeamName returns the winner's label (flag + 3-letter code) for a
// matchup, or "TBD" when the matchup is unresolved.
func nextRoundTeamName(mu api.WCMatchup) string {
	if mu.WinnerID == nil {
		return "TBD"
	}
	if *mu.WinnerID == mu.HomeTeamID {
		return MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
	}
	return MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway)
}

// renderBracketLine renders a single matchup line with flag emojis.
func renderBracketLine(mu api.WCMatchup) string {
	return renderBracketLineRaw(mu, true)
}

func renderBracketLineRaw(mu api.WCMatchup, showArrow bool) string {
	// Tie name column width to labelTargetWidth so any future change to the
	// label budget (see team_label.go) propagates here. The extra cells of
	// slack give the bracket some breathing room without affecting the
	// per-row width-invariance guarantee.
	const nameW = labelTargetWidth + 2
	const scoreW = 7

	home := MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
	away := MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway)

	homeIsWinner := mu.WinnerID != nil && *mu.WinnerID == mu.HomeTeamID
	awayIsWinner := mu.WinnerID != nil && *mu.WinnerID == mu.AwayTeamID

	var homeStr, awayStr string
	if homeIsWinner {
		homeStr = WinnerStyle.Width(nameW).Render(home)
	} else {
		homeStr = MatchLineStyle.Width(nameW).Render(home)
	}
	if awayIsWinner {
		awayStr = WinnerStyle.Width(nameW).Render(away)
	} else {
		awayStr = MatchLineStyle.Width(nameW).Render(away)
	}

	var scoreStr string
	if mu.HomeScore != nil && mu.AwayScore != nil {
		scoreStr = ScoreStyle.Render(fmt.Sprintf("%d–%d", *mu.HomeScore, *mu.AwayScore))
	} else {
		scoreStr = MatchLineStyle.Render(" vs ")
	}
	scoreStr = lipgloss.NewStyle().Width(scoreW).Align(lipgloss.Center).Render(scoreStr)

	if mu.WinnerID == nil || !showArrow {
		return MatchLineStyle.Render(fmt.Sprintf("  %s  %s  %s", homeStr, scoreStr, awayStr))
	}

	winnerLabel := MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
	if *mu.WinnerID == mu.AwayTeamID {
		winnerLabel = MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway)
	}
	penStr := ""
	if mu.IsPenalties {
		penStr = " " + PenStyle.Render("(p)")
	}
	winner := WinnerStyle.Render(winnerLabel) + penStr

	return fmt.Sprintf("  %s  %s  %s  %s",
		homeStr, scoreStr, awayStr, winner)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

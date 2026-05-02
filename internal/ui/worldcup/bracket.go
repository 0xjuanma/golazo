package worldcup

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// BracketLineCount returns the total number of content lines the bracket
// view produces. Must stay in sync with RenderBracket's line construction.
func BracketLineCount(wcData *api.WorldCupData) int {
	if wcData == nil {
		return 0
	}
	count := 0
	for i, round := range wcData.KnockoutRounds {
		count += 2 + 1 // roundHdr + blank + trailing blank
		count += renderBracketRoundLineCount(round, i)
	}
	if wcData.BronzeFinal != nil {
		count += 4 // bronzeHdr + blank + matchup + trailing blank
	}
	if wcData.Champion != nil {
		count += 2 // blank + champion line
	}
	return count
}

// renderBracketRoundLineCount returns the number of lines for a single round.
func renderBracketRoundLineCount(round api.WCKnockoutRound, roundIdx int) int {
	n := len(round.Matchups)
	// Paired connector lines: every two matchups share a connector (3 lines)
	pairs := n / 2
	singles := n % 2
	return n + pairs*3 + singles
}

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
	help := HelpStyle.Width(width).Render("j/k: scroll  Esc: back to groups  q: quit")

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
		emoji := FlagEmoji(wcData.Champion.ShortName)
		champ := ChampionStyle.Render(fmt.Sprintf("  🏆  Champion: %s %s", emoji, wcData.Champion.Name))
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

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderBracketRound renders all matchups in a round, pairing consecutive
// matchups with box-drawing connectors to show next-round opponents.
func renderBracketRound(round api.WCKnockoutRound, width int) []string {
	var lines []string
	mus := round.Matchups

	for i := 0; i < len(mus); i += 2 {
		mu1 := mus[i]
		lines = append(lines, renderBracketLine(mu1))

		if i+1 < len(mus) {
			mu2 := mus[i+1]

			// Connector: ──╮ / ├─► next / ──╯
			winnerLabel := ""
			if mu1.WinnerID != nil && mu2.WinnerID != nil {
				w1 := nextRoundTeamName(mu1)
				w2 := nextRoundTeamName(mu2)
				winnerLabel = fmt.Sprintf(" ► %s vs %s", w1, w2)
			}
			connector := ConnectorStyle.Render("──╮")
			middle := ConnectorStyle.Render("  ├─") + lipgloss.NewStyle().Foreground(colorDim).Render(winnerLabel)
			bottom := ConnectorStyle.Render("──╯")

			// Align connectors after the match line
			const matchLineW = 44 // nameW*2 + scoreW + spacing
			pad := strings.Repeat(" ", 2)
			lines = append(lines,
				pad+renderBracketLineRaw(mu2, false)+pad+connector,
				pad+strings.Repeat(" ", matchLineW)+middle,
				pad+renderBracketLine(mu1)[:0]+"  "+strings.Repeat(" ", matchLineW)+bottom,
			)
		}
	}
	return lines
}

// nextRoundTeamName returns the winner's short name for a matchup.
func nextRoundTeamName(mu api.WCMatchup) string {
	if mu.WinnerID == nil {
		return "TBD"
	}
	if *mu.WinnerID == mu.HomeTeamID {
		if mu.HomeShort != "" {
			return mu.HomeShort
		}
		return mu.HomeTeam
	}
	if mu.AwayShort != "" {
		return mu.AwayShort
	}
	return mu.AwayTeam
}

// renderBracketLine renders a single matchup line with flag emojis.
func renderBracketLine(mu api.WCMatchup) string {
	return renderBracketLineRaw(mu, true)
}

func renderBracketLineRaw(mu api.WCMatchup, showArrow bool) string {
	const nameW = 14
	const scoreW = 7

	home := teamDisplay(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
	away := teamDisplay(mu.AwayShort, mu.AwayTeam, mu.TBDAway)

	homeEmoji := FlagEmoji(mu.HomeShort)
	awayEmoji := FlagEmoji(mu.AwayShort)

	if len(home) > nameW {
		home = home[:nameW]
	}
	if len(away) > nameW {
		away = away[:nameW]
	}

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

	homeFlag := lipgloss.NewStyle().Width(3).Render(homeEmoji)
	awayFlag := lipgloss.NewStyle().Width(3).Render(awayEmoji)

	if mu.WinnerID == nil || !showArrow {
		return MatchLineStyle.Render(fmt.Sprintf("  %s%s  %s  %s%s", homeFlag, homeStr, scoreStr, awayFlag, awayStr))
	}

	winnerShort := mu.HomeShort
	if *mu.WinnerID == mu.AwayTeamID {
		winnerShort = mu.AwayShort
	}
	penStr := ""
	if mu.IsPenalties {
		penStr = " " + PenStyle.Render("(p)")
	}
	emoji := FlagEmoji(winnerShort)
	winner := WinnerStyle.Render(winnerShort) + penStr
	arrow := MatchLineStyle.Render("  ──► ")

	return fmt.Sprintf("  %s%s  %s  %s%s%s%s %s",
		homeFlag, homeStr, scoreStr, awayFlag, awayStr, arrow, emoji, winner)
}

// teamDisplay returns the display name for a team slot.
func teamDisplay(short, full string, tbd bool) string {
	if tbd {
		return "TBD"
	}
	if short != "" {
		return short
	}
	return full
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// ── List item ────────────────────────────────────────────────────────────────

// WCGroupItem is a bubbles list.Item for a single World Cup group.
// Exported so the app package can construct items when data loads.
type WCGroupItem struct {
	Group api.WCGroup
}

func (i WCGroupItem) FilterValue() string { return i.Group.Name }
func (i WCGroupItem) Title() string       { return i.Group.Name }
func (i WCGroupItem) Description() string {
	parts := make([]string, 0, len(i.Group.Teams))
	for _, t := range i.Group.Teams {
		name := t.Team.ShortName
		if name == "" {
			name = t.Team.Name
		}
		if len(name) > 3 {
			name = name[:3]
		}
		parts = append(parts, fmt.Sprintf("%s %d", name, t.Points))
	}
	return strings.Join(parts, "  ")
}

// NewWCGroupDelegate creates a styled list delegate for WC group items.
// Uses the same neon theme as the rest of the app.
func NewWCGroupDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetHeight(2) // title + description

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

	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(neonWhite).
		Padding(0, 1)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)

	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)

	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(neonDim).
		Padding(0, 1)

	return d
}

// ── Styles ───────────────────────────────────────────────────────────────────

var (
	wcLoadingStyle = lipgloss.NewStyle().
			Foreground(neonDim).
			Italic(true)

	wcErrorStyle = lipgloss.NewStyle().
			Foreground(neonRed)

	wcHelpStyle = lipgloss.NewStyle().
			Foreground(neonDim).
			Align(lipgloss.Center)

	wcPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(neonRed).
			Padding(0, 1)

	wcGroupHeaderStyle = lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true)

	wcQualifiedStyle = lipgloss.NewStyle().
				Foreground(neonCyan)

	wcEliminatedStyle = lipgloss.NewStyle().
				Foreground(neonDim)

	wcRoundHeaderStyle = lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true)

	wcMatchLineStyle = lipgloss.NewStyle().
				Foreground(neonWhite)

	wcWinnerStyle = lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)

	wcScoreStyle = lipgloss.NewStyle().
			Foreground(neonRed).
			Bold(true)

	wcPenStyle = lipgloss.NewStyle().
			Foreground(neonDim).
			Italic(true)
)

// ── Groups list view ─────────────────────────────────────────────────────────

// RenderWorldCupGroups renders the groups list using a bubbles/list component.
func RenderWorldCupGroups(width, height int, wcData *api.WorldCupData, groupsList list.Model, loading bool, lastErr string, bannerType constants.StatusBannerType) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	if loading {
		content := wcLoadingStyle.Render("Loading World Cup data...")
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, statusBanner, content))
	}

	if lastErr != "" {
		content := wcErrorStyle.Render(lastErr)
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, statusBanner, content))
	}

	if wcData == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			wcLoadingStyle.Render("No data"))
	}

	header := design.RenderHeader(wcData.Name+" — Groups", width-2)
	tabHint := lipgloss.NewStyle().Foreground(neonDim).Render("  b: Knockout Bracket")
	help := wcHelpStyle.Width(width).Render("↑/↓: navigate  Enter: group detail  b: bracket  /: filter  Esc: back  q: quit")

	overhead := 4
	if statusBanner != "" {
		overhead++
	}
	listHeight := height - overhead
	if listHeight < 4 {
		listHeight = 4
	}
	groupsList.SetSize(width, listHeight)

	return lipgloss.JoinVertical(lipgloss.Left,
		statusBanner,
		header,
		tabHint,
		"",
		groupsList.View(),
		help,
	)
}

// ── Group detail view ─────────────────────────────────────────────────────────

// RenderWorldCupGroupDetail renders the expanded standings for a single group.
func RenderWorldCupGroupDetail(width, height int, wcData *api.WorldCupData, groupIdx int, bannerType constants.StatusBannerType) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	if wcData == nil {
		return wcLoadingStyle.Render("Loading group data...")
	}
	if groupIdx < 0 || groupIdx >= len(wcData.Groups) {
		return wcErrorStyle.Render(fmt.Sprintf("Group index %d out of range", groupIdx))
	}

	g := wcData.Groups[groupIdx]
	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	header := design.RenderHeader(wcData.Name+" — "+g.Name, width-2)

	// Standings table — full width, no Height() constraint
	tableContent := renderWCStandingsTable(g, width-4)
	table := wcPanelStyle.Width(width - 2).Render(tableContent)

	// Qualification row beneath the table
	qual := renderWCQualificationRow(g, width)

	help := wcHelpStyle.Width(width).Render("Esc: back to groups  q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		statusBanner,
		header,
		"",
		table,
		"",
		qual,
		"",
		help,
	)
}

// renderWCStandingsTable renders the group standings table at the given width.
func renderWCStandingsTable(g api.WCGroup, width int) string {
	if len(g.Teams) == 0 {
		return wcLoadingStyle.Render("No standings data")
	}

	// Column widths: # (3) + sp (2) + Team (nameW) + P(4) + W(4) + D(4) + L(4) + GF(4) + GA(4) + GD(5) + Pts(4) = nameW + 38
	nameW := width - 38
	if nameW < 8 {
		nameW = 8
	}
	if nameW > 20 {
		nameW = 20
	}

	hdr := lipgloss.JoinHorizontal(lipgloss.Top,
		dialogHeaderStyle.Width(3).Align(lipgloss.Right).Render("#"),
		"  ",
		dialogHeaderStyle.Width(nameW).Render("Team"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("P"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("W"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("D"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("L"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("GF"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("GA"),
		dialogHeaderStyle.Width(5).Align(lipgloss.Right).Render("GD"),
		dialogHeaderStyle.Width(4).Align(lipgloss.Right).Render("Pts"),
	)

	sepWidth := 3 + 2 + nameW + 4 + 4 + 4 + 4 + 4 + 4 + 5 + 4
	sep := dialogSeparatorStyle.Render(strings.Repeat("─", sepWidth))

	lines := []string{hdr, sep}
	for i, t := range g.Teams {
		name := t.Team.Name
		if len(name) > nameW {
			name = name[:nameW-1] + "…"
		}

		isQualified := i < 2
		teamStyle := wcEliminatedStyle
		ptsStyle := wcEliminatedStyle
		if isQualified {
			teamStyle = wcQualifiedStyle
			ptsStyle = lipgloss.NewStyle().Foreground(neonCyan).Bold(true)
		}

		gdStr := fmt.Sprintf("%+d", t.GoalDifference)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			dialogAlignRight(3, fmt.Sprintf("%d", t.Position)),
			"  ",
			teamStyle.Width(nameW).Render(name),
			dialogAlignRight(4, fmt.Sprintf("%d", t.Played)),
			dialogAlignRight(4, fmt.Sprintf("%d", t.Won)),
			dialogAlignRight(4, fmt.Sprintf("%d", t.Drawn)),
			dialogAlignRight(4, fmt.Sprintf("%d", t.Lost)),
			dialogAlignRight(4, fmt.Sprintf("%d", t.GoalsFor)),
			dialogAlignRight(4, fmt.Sprintf("%d", t.GoalsAgainst)),
			dialogAlignRight(5, gdStr),
			ptsStyle.Width(4).Align(lipgloss.Right).Render(fmt.Sprintf("%d", t.Points)),
		)
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
}

// renderWCQualificationRow renders a compact one-line qualification summary.
func renderWCQualificationRow(g api.WCGroup, width int) string {
	var parts []string
	for i, t := range g.Teams {
		name := t.Team.Name
		if i < 2 {
			parts = append(parts, wcQualifiedStyle.Render("✓ "+name))
		} else {
			parts = append(parts, wcEliminatedStyle.Render("✗ "+name))
		}
	}
	line := strings.Join(parts, "   ")
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(line)
}

// ── Bracket view ─────────────────────────────────────────────────────────────

// RenderWorldCupBracket renders the knockout bracket as an enhanced vertical list.
// Each round is shown as a section. Winners are shown with ──► arrows.
// Indentation increases with each round to visually convey progression.
func RenderWorldCupBracket(width, height int, wcData *api.WorldCupData, scrollOffset int, bannerType constants.StatusBannerType) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	if wcData == nil {
		return wcLoadingStyle.Render("No bracket data")
	}

	statusBanner := renderStatusBanner(bannerType, width)
	if statusBanner != "" {
		statusBanner += "\n"
	}

	header := design.RenderHeader(wcData.Name+" — Knockout Bracket", width-2)
	help := wcHelpStyle.Width(width).Render("j/k: scroll  Esc: back to groups  q: quit")

	// Build content lines
	var lines []string
	indentPerRound := map[string]int{
		"1/32": 0, "1/16": 0, "1/8": 2, "1/4": 4, "1/2": 6, "final": 8,
	}

	for _, round := range wcData.KnockoutRounds {
		indent := indentPerRound[round.Stage]
		pad := strings.Repeat(" ", indent)

		roundHdr := wcRoundHeaderStyle.Render("── " + strings.ToUpper(round.Label) + " " + strings.Repeat("─", max(0, width-len(round.Label)-8)))
		lines = append(lines, roundHdr, "")

		for _, mu := range round.Matchups {
			lines = append(lines, pad+renderWCBracketLine(mu))
		}
		lines = append(lines, "")
	}

	// Bronze final
	if wcData.BronzeFinal != nil {
		bronzeHdr := wcRoundHeaderStyle.Render("── 3RD PLACE " + strings.Repeat("─", max(0, width-17)))
		lines = append(lines, bronzeHdr, "")
		lines = append(lines, renderWCBracketLine(*wcData.BronzeFinal))
		lines = append(lines, "")
	}

	// Champion line
	if wcData.Champion != nil {
		lines = append(lines, "")
		champion := lipgloss.NewStyle().Foreground(neonCyan).Bold(true).
			Render(fmt.Sprintf("  🏆  Champion: %s", wcData.Champion.Name))
		lines = append(lines, champion)
	}

	// Scroll window — overhead: statusBanner(0-1) + header(1) + gap(1) + panel borders(2) + gap(1) + help(1) = 6-7
	overhead := 7
	if statusBanner == "" {
		overhead = 6
	}
	availableLines := height - overhead
	if availableLines < 1 {
		availableLines = 1
	}

	// Clamp scroll
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

	// Scroll indicator
	scrollIndicator := ""
	if len(lines) > availableLines {
		scrollIndicator = lipgloss.NewStyle().Foreground(neonDim).
			Render(fmt.Sprintf("  (%d/%d lines)", scrollOffset+1, len(lines)))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		statusBanner,
		header,
		"",
		wcPanelStyle.Width(width-2).Render(visible),
		scrollIndicator,
		help,
	)
}

// renderWCBracketLine renders a single matchup line:
//
//	Home  HS–AS  Away        ──►  Winner (pens)
func renderWCBracketLine(mu api.WCMatchup) string {
	const nameW = 16
	const scoreW = 7

	home := mu.HomeShort
	if home == "" {
		home = mu.HomeTeam
	}
	if mu.TBDHome {
		home = "TBD"
	}
	if len(home) > nameW {
		home = home[:nameW]
	}

	away := mu.AwayShort
	if away == "" {
		away = mu.AwayTeam
	}
	if mu.TBDAway {
		away = "TBD"
	}
	if len(away) > nameW {
		away = away[:nameW]
	}

	homeIsWinner := mu.WinnerID != nil && *mu.WinnerID == mu.HomeTeamID
	awayIsWinner := mu.WinnerID != nil && *mu.WinnerID == mu.AwayTeamID

	homeStr := lipgloss.NewStyle().Width(nameW).Render(home)
	if homeIsWinner {
		homeStr = wcWinnerStyle.Width(nameW).Render(home)
	}
	awayStr := lipgloss.NewStyle().Width(nameW).Render(away)
	if awayIsWinner {
		awayStr = wcWinnerStyle.Width(nameW).Render(away)
	}

	var scoreStr string
	if mu.HomeScore != nil && mu.AwayScore != nil {
		scoreStr = wcScoreStyle.Render(fmt.Sprintf("%d – %d", *mu.HomeScore, *mu.AwayScore))
	} else {
		scoreStr = wcMatchLineStyle.Render("  vs ")
	}
	// Pad score to consistent visual width
	scoreStr = lipgloss.NewStyle().Width(scoreW).Align(lipgloss.Center).Render(scoreStr)

	if mu.WinnerID == nil {
		return wcMatchLineStyle.Render(fmt.Sprintf("  %s  %s  %s", homeStr, scoreStr, awayStr))
	}

	winnerID := *mu.WinnerID
	winnerName := mu.HomeTeam
	if winnerID == mu.AwayTeamID {
		winnerName = mu.AwayTeam
	}
	winnerShort := mu.HomeShort
	if winnerID == mu.AwayTeamID {
		winnerShort = mu.AwayShort
	}
	if winnerShort == "" {
		winnerShort = winnerName
	}

	penStr := ""
	if mu.IsPenalties {
		penStr = " " + wcPenStyle.Render("(p)")
	}

	arrow := wcMatchLineStyle.Render("  ──► ")
	winner := wcWinnerStyle.Render(winnerShort) + penStr

	return fmt.Sprintf("  %s  %s  %s%s%s",
		homeStr, scoreStr, awayStr, arrow, winner)
}

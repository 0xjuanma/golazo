package worldcup

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// ── List item ────────────────────────────────────────────────────────────────

// WCGroupItem is a bubbles list.Item for a single World Cup group.
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
		emoji := FlagEmoji(t.Team.ShortName)
		if emoji != "" {
			parts = append(parts, fmt.Sprintf("%s %s %d", emoji, name, t.Points))
		} else {
			parts = append(parts, fmt.Sprintf("%s %d", name, t.Points))
		}
	}
	return strings.Join(parts, "  ")
}

// NewWCGroupDelegate creates a styled list delegate for WC group items.
func NewWCGroupDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetHeight(2)

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(colorRed).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorRed)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(colorCyan).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorRed)

	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(colorWhite).
		Padding(0, 1)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(colorDim).
		Padding(0, 1)

	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(colorDim).
		Padding(0, 1)

	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(colorDim).
		Padding(0, 1)

	return d
}

// ── Groups list view ─────────────────────────────────────────────────────────

// RenderGroupsList renders the groups overview using a bubbles/list component.
func RenderGroupsList(width, height int, wcData *api.WorldCupData, groupsList list.Model, loading bool, lastErr, statusBanner string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	if loading {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			LoadingStyle.Render("Loading World Cup data..."))
	}
	if lastErr != "" {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			ErrorStyle.Render(lastErr))
	}
	if wcData == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			LoadingStyle.Render("No data"))
	}

	titleText := wcData.Name + " — Groups"
	header := design.RenderHeader(titleText, width-2)

	phase := DerivePhase(wcData)
	phaseHint := ""
	if phase != "" {
		phaseHint = lipgloss.NewStyle().Foreground(colorGold).Render("  " + phase)
	}

	keyHints := []string{"  b: Knockout Bracket", "  g: Groups Grid"}
	tabHint := lipgloss.NewStyle().Foreground(colorDim).Render(strings.Join(keyHints, "  "))

	help := HelpStyle.Width(width).Render("↑/↓: navigate  Enter: detail  b: bracket  g: grid  /: filter  Esc: back  q: quit")

	overhead := 4
	if statusBanner != "" {
		overhead++
	}
	listHeight := height - overhead
	if listHeight < 4 {
		listHeight = 4
	}
	groupsList.SetSize(width, listHeight)

	parts := []string{}
	if statusBanner != "" {
		parts = append(parts, statusBanner)
	}
	parts = append(parts, header)
	if phaseHint != "" {
		parts = append(parts, phaseHint)
	}
	parts = append(parts, tabHint, "", groupsList.View(), help)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ── Group detail view ─────────────────────────────────────────────────────────

// RenderGroupDetail renders the expanded standings for a single group,
// including a pixel flag for each team when available.
func RenderGroupDetail(width, height int, wcData *api.WorldCupData, groupIdx int, statusBanner string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if wcData == nil {
		return LoadingStyle.Render("Loading group data...")
	}
	if groupIdx < 0 || groupIdx >= len(wcData.Groups) {
		return ErrorStyle.Render(fmt.Sprintf("Group index %d out of range", groupIdx))
	}

	g := wcData.Groups[groupIdx]

	header := design.RenderHeader(wcData.Name+" — "+g.Name, width-2)
	tableContent := renderGroupStandingsTable(g, width-4)
	table := PanelStyle.Width(width - 2).Render(tableContent)
	qual := renderQualificationRow(g, width)
	help := HelpStyle.Width(width).Render("Esc: back to groups  q: quit")

	parts := []string{}
	if statusBanner != "" {
		parts = append(parts, statusBanner)
	}
	parts = append(parts, header, "", table, "", qual, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderGroupStandingsTable renders the full standings table for a group.
// Teams are styled: top 2 = cyan (qualified), 3rd = gold (possible), 4th = dim.
// Flag emojis precede team names.
func renderGroupStandingsTable(g api.WCGroup, width int) string {
	if len(g.Teams) == 0 {
		return LoadingStyle.Render("No standings data")
	}

	// Col widths: # (3) + 2sp + emoji (2) + 1sp + Team (nameW) + P(4)+W(4)+D(4)+L(4)+GF(4)+GA(4)+GD(5)+Pts(4) = nameW + 42
	const emojiW = 3
	nameW := width - 42 - emojiW
	if nameW < 8 {
		nameW = 8
	}
	if nameW > 20 {
		nameW = 20
	}

	hdr := lipgloss.JoinHorizontal(lipgloss.Top,
		HeaderStyle.Width(3).Align(lipgloss.Right).Render("#"),
		"  ",
		HeaderStyle.Width(emojiW).Render(""),
		HeaderStyle.Width(nameW).Render("Team"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("P"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("W"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("D"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("L"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("GF"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("GA"),
		HeaderStyle.Width(5).Align(lipgloss.Right).Render("GD"),
		HeaderStyle.Width(4).Align(lipgloss.Right).Render("Pts"),
	)

	sepWidth := 3 + 2 + emojiW + nameW + 4 + 4 + 4 + 4 + 4 + 4 + 5 + 4
	sep := SepStyle.Render(strings.Repeat("─", sepWidth))

	lines := []string{hdr, sep}
	for i, t := range g.Teams {
		name := t.Team.Name
		if len(name) > nameW {
			name = name[:nameW-1] + "…"
		}

		emoji := FlagEmoji(t.Team.ShortName)

		var teamStyle, ptsStyle lipgloss.Style
		switch {
		case i < 2:
			teamStyle = QualifiedStyle
			ptsStyle = QualPtsStyle
		case i == 2:
			teamStyle = ThirdStyle
			ptsStyle = ThirdPtsStyle
		default:
			teamStyle = EliminatedStyle
			ptsStyle = EliminPtsStyle
		}

		gdStr := fmt.Sprintf("%+d", t.GoalDifference)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			alignRight(3, fmt.Sprintf("%d", t.Position)),
			"  ",
			lipgloss.NewStyle().Width(emojiW).Render(emoji),
			teamStyle.Width(nameW).Render(name),
			alignRight(4, fmt.Sprintf("%d", t.Played)),
			alignRight(4, fmt.Sprintf("%d", t.Won)),
			alignRight(4, fmt.Sprintf("%d", t.Drawn)),
			alignRight(4, fmt.Sprintf("%d", t.Lost)),
			alignRight(4, fmt.Sprintf("%d", t.GoalsFor)),
			alignRight(4, fmt.Sprintf("%d", t.GoalsAgainst)),
			alignRight(5, gdStr),
			ptsStyle.Width(4).Align(lipgloss.Right).Render(fmt.Sprintf("%d", t.Points)),
		)
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
}

// renderQualificationRow renders a compact one-line summary of qualified teams.
func renderQualificationRow(g api.WCGroup, width int) string {
	var parts []string
	for i, t := range g.Teams {
		name := t.Team.Name
		emoji := FlagEmoji(t.Team.ShortName)
		display := name
		if emoji != "" {
			display = emoji + " " + name
		}
		switch {
		case i < 2:
			parts = append(parts, QualifiedStyle.Render("✓ "+display))
		case i == 2:
			parts = append(parts, ThirdStyle.Render("? "+display))
		default:
			parts = append(parts, EliminatedStyle.Render("✗ "+display))
		}
	}
	line := strings.Join(parts, "   ")
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(line)
}

// ── Groups grid view ─────────────────────────────────────────────────────────

// RenderGroupGrid renders all groups in a compact grid overview (2 per row).
// selectedGroupIdx highlights the currently selected group cell.
func RenderGroupGrid(width, height int, wcData *api.WorldCupData, selectedGroupIdx int, statusBanner string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if wcData == nil {
		return LoadingStyle.Render("No data")
	}

	header := design.RenderHeader(wcData.Name+" — Groups Overview", width-2)
	help := HelpStyle.Width(width).Render("↑/↓/←/→: navigate  Enter: detail  Esc: back  b: bracket  q: quit")

	cols := 2
	if width > 120 {
		cols = 4
	} else if width > 80 {
		cols = 3
	}

	cellW := (width - cols - 2) / cols
	if cellW < 20 {
		cellW = 20
	}

	var rows []string
	for rowStart := 0; rowStart < len(wcData.Groups); rowStart += cols {
		var cells []string
		for c := 0; c < cols; c++ {
			gIdx := rowStart + c
			if gIdx >= len(wcData.Groups) {
				cells = append(cells, lipgloss.NewStyle().Width(cellW+2).Render(""))
				continue
			}
			g := wcData.Groups[gIdx]
			selected := gIdx == selectedGroupIdx

			content := renderGroupGridCell(g, cellW-4)

			var cellStyle lipgloss.Style
			if selected {
				cellStyle = GridSelectedGroupStyle.Width(cellW)
			} else {
				cellStyle = GridNormalGroupStyle.Width(cellW)
			}
			cells = append(cells, cellStyle.Render(content))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	gridContent := strings.Join(rows, "\n")

	parts := []string{}
	if statusBanner != "" {
		parts = append(parts, statusBanner)
	}
	parts = append(parts, header, "", gridContent, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderGroupGridCell renders a mini standings table for a single group cell.
func renderGroupGridCell(g api.WCGroup, width int) string {
	title := GridGroupHeaderStyle.Render(g.Name)
	lines := []string{title}
	for i, t := range g.Teams {
		name := t.Team.ShortName
		if name == "" {
			name = t.Team.Name
		}
		if len(name) > 3 {
			name = name[:3]
		}
		emoji := FlagEmoji(t.Team.ShortName)

		pts := fmt.Sprintf("%2d", t.Points)

		var ts lipgloss.Style
		switch {
		case i < 2:
			ts = GridTeamQualStyle
		case i == 2:
			ts = GridTeamThirdStyle
		default:
			ts = GridTeamDimStyle
		}

		line := ts.Render(fmt.Sprintf("  %s %-3s %s", emoji, name, pts))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

package worldcup

import (
	"fmt"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/charmbracelet/lipgloss"
)

// RenderTrophyCard renders a prominent champion + runner-up card
// with pixel flags when available, suitable for the bottom of the bracket view.
func RenderTrophyCard(wcData *api.WorldCupData, width int) string {
	if wcData == nil || wcData.Champion == nil {
		return ""
	}

	champEmoji := FlagEmoji(wcData.Champion.ShortName)
	champName := wcData.Champion.Name
	champFlag := RenderPixelFlag(wcData.Champion.ShortName)

	var runnerUpBlock string
	if wcData.RunnerUp != nil {
		ruEmoji := FlagEmoji(wcData.RunnerUp.ShortName)
		ruName := wcData.RunnerUp.Name
		ruFlag := RenderPixelFlag(wcData.RunnerUp.ShortName)

		champCol := buildFlagNameBlock(champFlag, champEmoji, champName, "🏆 Champion", true)
		ruCol := buildFlagNameBlock(ruFlag, ruEmoji, ruName, "Runner-up", false)

		runnerUpBlock = lipgloss.JoinHorizontal(lipgloss.Center,
			champCol,
			lipgloss.NewStyle().Width(4).Render(""),
			ruCol,
		)
	} else {
		block := buildFlagNameBlock(champFlag, champEmoji, champName, "🏆 Champion", true)
		runnerUpBlock = block
	}

	card := TrophyCardStyle.Render(runnerUpBlock)
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(card)
}

// buildFlagNameBlock creates a vertically stacked block: pixel flag + label + name.
func buildFlagNameBlock(pixelFlag, emoji, name, label string, isChamp bool) string {
	var labelStyle lipgloss.Style
	if isChamp {
		labelStyle = ChampionStyle
	} else {
		labelStyle = lipgloss.NewStyle().Foreground(colorDim).Bold(true)
	}

	lines := []string{}
	if pixelFlag != "" {
		lines = append(lines, pixelFlag)
	}
	lines = append(lines,
		labelStyle.Render(label),
		lipgloss.NewStyle().Foreground(colorWhite).Bold(isChamp).
			Render(fmt.Sprintf("%s %s", emoji, name)),
	)
	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}

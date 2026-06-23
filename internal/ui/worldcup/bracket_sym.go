package worldcup

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderSymmetricBracket renders the knockout bracket in a symmetric two-tab layout.
// tab 0 = outer rounds (R32 list or R16 pairs). tab 1 = inner rounds (QF/SF/Final tree).
func RenderSymmetricBracket(width, height int, wcData *api.WorldCupData, tab int, banner string) string {
	if width <= 0 {
		width = 80
	}
	if wcData == nil {
		return LoadingStyle.Render("No bracket data")
	}

	header := design.RenderHeader(wcData.Name+" — Knockout Bracket", width-2)
	tabBar := symTabBar(tab, wcData)
	help := HelpStyle.Width(width).Render("← →: tab  u: upcoming  esc: back  q: quit")

	var body string
	switch tab {
	case 0:
		body = symTab0(wcData)
	default:
		body = symTab1(wcData)
	}

	parts := []string{}
	if banner != "" {
		parts = append(parts, banner)
	}
	parts = append(parts, header, "", tabBar, "", body, "", help)
	return padToHeight(lipgloss.JoinVertical(lipgloss.Left, parts...), height)
}

func symTabBar(tab int, wcData *api.WorldCupData) string {
	hasR32 := symRound(wcData.KnockoutRounds, "1/16") != nil
	t0, t1 := "R16", "QF › Final"
	if hasR32 {
		t0, t1 = "R32 | R16", "R16 › Final"
	}
	act := lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	dim := lipgloss.NewStyle().Foreground(colorDim)
	sep := dim.Render("  ·  ")
	if tab == 0 {
		return "  " + act.Render("["+t0+"]") + sep + dim.Render("["+t1+"]")
	}
	return "  " + dim.Render("["+t0+"]") + sep + act.Render("["+t1+"]")
}

func symRound(rounds []api.WCKnockoutRound, stage string) *api.WCKnockoutRound {
	for i := range rounds {
		if rounds[i].Stage == stage {
			return &rounds[i]
		}
	}
	return nil
}

// ─── Tab 0 ───────────────────────────────────────────────────────────────────

func symTab0(wcData *api.WorldCupData) string {
	r32 := symRound(wcData.KnockoutRounds, "1/16")
	r16 := symRound(wcData.KnockoutRounds, "1/8")
	if r32 != nil && len(r32.Matchups) >= 16 {
		return symR32List(r32.Matchups)
	}
	if r16 != nil && len(r16.Matchups) >= 8 {
		return symR16Pairs(r16.Matchups)
	}
	return LoadingStyle.Render("Bracket data not yet available")
}

// symR32List renders R32 as a compact 2-column list grouped in pairs.
// Left col: R32[0..7], right col: R32[8..15]. Groups of 2 share an R16 slot.
func symR32List(r32 []api.WCMatchup) string {
	var lines []string
	for i := 0; i < 4; i++ {
		l0, l1 := r32[2*i], r32[2*i+1]
		r0, r1 := r32[2*i+8], r32[2*i+9]
		lines = append(lines,
			symSideBySide(symCompact(l0), symCompact(r0)),
			symSideBySide(symCompact(l1), symCompact(r1)),
		)
		if i < 3 {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

const symColW = 22 // visual width of each half-line column

func symSideBySide(left, right string) string {
	lw := lipgloss.Width(left)
	pad := ""
	if lw < symColW {
		pad = strings.Repeat(" ", symColW-lw)
	}
	return left + pad + "    " + right
}

func symCompact(mu api.WCMatchup) string {
	home := MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
	away := MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway)
	hW := mu.WinnerID != nil && *mu.WinnerID == mu.HomeTeamID
	aW := mu.WinnerID != nil && *mu.WinnerID == mu.AwayTeamID
	hs, as_ := MatchLineStyle, MatchLineStyle
	if hW {
		hs = WinnerStyle
	}
	if aW {
		as_ = WinnerStyle
	}
	var score string
	if mu.HomeScore != nil && mu.AwayScore != nil {
		score = ScoreStyle.Render(fmt.Sprintf("%d–%d", *mu.HomeScore, *mu.AwayScore))
		if mu.IsPenalties {
			score += PenStyle.Render("p")
		}
	} else {
		score = MatchLineStyle.Render("vs")
	}
	return hs.Render(home) + " " + score + " " + as_.Render(away)
}

// symR16Pairs renders R16 as symmetric 3-line triplets (for tournaments without R32).
func symR16Pairs(r16 []api.WCMatchup) string {
	var lines []string
	for i := 0; i < 4; i++ {
		lines = append(lines, symTriplet(symGet(r16, i), symGet(r16, i+4))...)
		if i < 3 {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

// symTriplet renders one left matchup and one right matchup on 3 symmetric lines.
func symTriplet(lmu, rmu api.WCMatchup) []string {
	lh := MatchupTeamLabel(lmu.HomeShort, lmu.HomeTeam, lmu.TBDHome)
	la := MatchupTeamLabel(lmu.AwayShort, lmu.AwayTeam, lmu.TBDAway)
	lw := nextRoundTeamName(lmu)
	rh := MatchupTeamLabel(rmu.HomeShort, rmu.HomeTeam, rmu.TBDHome)
	ra := MatchupTeamLabel(rmu.AwayShort, rmu.AwayTeam, rmu.TBDAway)
	rw := nextRoundTeamName(rmu)

	c := ConnectorStyle.Render
	t := MatchLineStyle.Render
	w := WinnerStyle.Render
	indent := strings.Repeat(" ", labelTargetWidth+3)
	gap := "     "

	l0 := t(lh) + c(" ─┐")
	l1 := indent + c("├─ ") + w(lw)
	l2 := t(la) + c(" ─┘")
	r0 := c("┌─ ") + t(rh)
	r1 := w(rw) + c(" ─┤")
	r2 := c("└─ ") + t(ra)

	mw := lipgloss.Width(l0)
	if lipgloss.Width(l1) > mw {
		mw = lipgloss.Width(l1)
	}
	if lipgloss.Width(l2) > mw {
		mw = lipgloss.Width(l2)
	}
	pad := func(s string) string {
		p := mw - lipgloss.Width(s)
		if p > 0 {
			return s + strings.Repeat(" ", p)
		}
		return s
	}
	return []string{pad(l0) + gap + r0, pad(l1) + gap + r1, pad(l2) + gap + r2}
}

// ─── Tab 1 ───────────────────────────────────────────────────────────────────

func symTab1(wcData *api.WorldCupData) string {
	hasR32 := symRound(wcData.KnockoutRounds, "1/16") != nil
	qf := symRound(wcData.KnockoutRounds, "1/4")
	sf := symRound(wcData.KnockoutRounds, "1/2")
	fin := symRound(wcData.KnockoutRounds, "final")
	if sf == nil || fin == nil || qf == nil {
		return LoadingStyle.Render("Bracket data not yet available")
	}
	if hasR32 {
		r16 := symRound(wcData.KnockoutRounds, "1/8")
		if r16 != nil {
			return sym4Level(r16.Matchups, qf.Matchups, sf.Matchups, fin.Matchups, wcData)
		}
	}
	return sym2Level(qf.Matchups, sf.Matchups, fin.Matchups, wcData)
}

// sym4Level builds a 15-line symmetric bracket for R16 → QF → SF → Final.
//
// Left side column positions (visual chars):
//
//	col0: R16 team label (0-8, 9 chars: label6 + " ─┐"3)
//	col1: QF connector (9-20: sp9 + "├─ " + label6 + " ─┐"3 = 21 chars)
//	sfCol: SF connector │ at position 20
func sym4Level(r16, qf, sf, fin []api.WCMatchup, wcData *api.WorldCupData) string {
	lH := [4]string{
		MatchupTeamLabel(symGet(r16, 0).HomeShort, symGet(r16, 0).HomeTeam, symGet(r16, 0).TBDHome),
		MatchupTeamLabel(symGet(r16, 1).HomeShort, symGet(r16, 1).HomeTeam, symGet(r16, 1).TBDHome),
		MatchupTeamLabel(symGet(r16, 2).HomeShort, symGet(r16, 2).HomeTeam, symGet(r16, 2).TBDHome),
		MatchupTeamLabel(symGet(r16, 3).HomeShort, symGet(r16, 3).HomeTeam, symGet(r16, 3).TBDHome),
	}
	lA := [4]string{
		MatchupTeamLabel(symGet(r16, 0).AwayShort, symGet(r16, 0).AwayTeam, symGet(r16, 0).TBDAway),
		MatchupTeamLabel(symGet(r16, 1).AwayShort, symGet(r16, 1).AwayTeam, symGet(r16, 1).TBDAway),
		MatchupTeamLabel(symGet(r16, 2).AwayShort, symGet(r16, 2).AwayTeam, symGet(r16, 2).TBDAway),
		MatchupTeamLabel(symGet(r16, 3).AwayShort, symGet(r16, 3).AwayTeam, symGet(r16, 3).TBDAway),
	}
	rH := [4]string{
		MatchupTeamLabel(symGet(r16, 4).HomeShort, symGet(r16, 4).HomeTeam, symGet(r16, 4).TBDHome),
		MatchupTeamLabel(symGet(r16, 5).HomeShort, symGet(r16, 5).HomeTeam, symGet(r16, 5).TBDHome),
		MatchupTeamLabel(symGet(r16, 6).HomeShort, symGet(r16, 6).HomeTeam, symGet(r16, 6).TBDHome),
		MatchupTeamLabel(symGet(r16, 7).HomeShort, symGet(r16, 7).HomeTeam, symGet(r16, 7).TBDHome),
	}
	rA := [4]string{
		MatchupTeamLabel(symGet(r16, 4).AwayShort, symGet(r16, 4).AwayTeam, symGet(r16, 4).TBDAway),
		MatchupTeamLabel(symGet(r16, 5).AwayShort, symGet(r16, 5).AwayTeam, symGet(r16, 5).TBDAway),
		MatchupTeamLabel(symGet(r16, 6).AwayShort, symGet(r16, 6).AwayTeam, symGet(r16, 6).TBDAway),
		MatchupTeamLabel(symGet(r16, 7).AwayShort, symGet(r16, 7).AwayTeam, symGet(r16, 7).TBDAway),
	}

	// R16 winners advance into QF slots
	lQF0h := nextRoundTeamName(symGet(r16, 0))
	lQF0a := nextRoundTeamName(symGet(r16, 1))
	lQF1h := nextRoundTeamName(symGet(r16, 2))
	lQF1a := nextRoundTeamName(symGet(r16, 3))
	rQF2h := nextRoundTeamName(symGet(r16, 4))
	rQF2a := nextRoundTeamName(symGet(r16, 5))
	rQF3h := nextRoundTeamName(symGet(r16, 6))
	rQF3a := nextRoundTeamName(symGet(r16, 7))

	// QF winners advance into SF slots
	lSFh := nextRoundTeamName(symGet(qf, 0))
	lSFa := nextRoundTeamName(symGet(qf, 1))
	rSFh := nextRoundTeamName(symGet(qf, 2))
	rSFa := nextRoundTeamName(symGet(qf, 3))

	c := ConnectorStyle.Render
	t := MatchLineStyle.Render
	w := WinnerStyle.Render

	// sfCol=20: │ at visual position 20 in left block.
	// After label(6)+" ─┐"(3)=9 chars: need 11 more spaces to reach col 20.
	qfSp := strings.Repeat(" ", 9)  // indent before QF ├─
	sfSp := strings.Repeat(" ", 20) // indent before SF ├─
	sfMid := strings.Repeat(" ", 11) // filler between R16 connector end (col9) and SF │ (col20)

	ll := make([]string, 15)
	ll[0] = t(lH[0]) + c(" ─┐")
	ll[1] = qfSp + c("├─ ") + w(lQF0h) + c(" ─┐")
	ll[2] = t(lA[0]) + c(" ─┘") + sfMid + c("│")
	ll[3] = sfSp + c("├─ ") + w(lSFh)
	ll[4] = t(lH[1]) + c(" ─┐") + sfMid + c("│")
	ll[5] = qfSp + c("├─ ") + w(lQF0a) + c(" ─┘")
	ll[6] = t(lA[1]) + c(" ─┘")
	ll[7] = sfSp + c("│")
	ll[8] = t(lH[2]) + c(" ─┐") + sfMid + c("│")
	ll[9] = qfSp + c("├─ ") + w(lQF1h) + c(" ─┐")
	ll[10] = t(lA[2]) + c(" ─┘") + sfMid + c("│")
	ll[11] = sfSp + c("├─ ") + w(lSFa)
	ll[12] = t(lH[3]) + c(" ─┐") + sfMid + c("│")
	ll[13] = qfSp + c("├─ ") + w(lQF1a) + c(" ─┘")
	ll[14] = t(lA[3]) + c(" ─┘")

	// Right side: SF connector │ at position 0, bracket opens leftward.
	// "┌─ " + team (R16 teams), w(label) + " ─┤" (QF winner node).
	rSp := strings.Repeat(" ", 9) // gap between SF │ and ┌─/└─ or QF label
	rr := make([]string, 15)
	rr[0] = " " + rSp + c("┌─ ") + t(rH[0])
	rr[1] = " " + w(rQF2h) + c(" ─┤")
	rr[2] = c("│") + rSp + c("└─ ") + t(rA[0])
	rr[3] = c("├─ ") + w(rSFh)
	rr[4] = c("│") + rSp + c("┌─ ") + t(rH[1])
	rr[5] = " " + w(rQF2a) + c(" ─┘")
	rr[6] = " " + rSp + c("└─ ") + t(rA[1])
	rr[7] = c("│")
	rr[8] = c("│") + rSp + c("┌─ ") + t(rH[2])
	rr[9] = " " + w(rQF3h) + c(" ─┤")
	rr[10] = c("│") + rSp + c("└─ ") + t(rA[2])
	rr[11] = c("├─ ") + w(rSFa)
	rr[12] = c("│") + rSp + c("┌─ ") + t(rH[3])
	rr[13] = " " + w(rQF3a) + c(" ─┘")
	rr[14] = " " + rSp + c("└─ ") + t(rA[3])

	return symJoin(ll, rr, 15, symFinalLabel(fin, wcData))
}

// sym2Level builds a 7-line symmetric bracket for QF → SF → Final (2022 format).
func sym2Level(qf, sf, fin []api.WCMatchup, wcData *api.WorldCupData) string {
	lH := [2]string{
		MatchupTeamLabel(symGet(qf, 0).HomeShort, symGet(qf, 0).HomeTeam, symGet(qf, 0).TBDHome),
		MatchupTeamLabel(symGet(qf, 1).HomeShort, symGet(qf, 1).HomeTeam, symGet(qf, 1).TBDHome),
	}
	lA := [2]string{
		MatchupTeamLabel(symGet(qf, 0).AwayShort, symGet(qf, 0).AwayTeam, symGet(qf, 0).TBDAway),
		MatchupTeamLabel(symGet(qf, 1).AwayShort, symGet(qf, 1).AwayTeam, symGet(qf, 1).TBDAway),
	}
	rH := [2]string{
		MatchupTeamLabel(symGet(qf, 2).HomeShort, symGet(qf, 2).HomeTeam, symGet(qf, 2).TBDHome),
		MatchupTeamLabel(symGet(qf, 3).HomeShort, symGet(qf, 3).HomeTeam, symGet(qf, 3).TBDHome),
	}
	rA := [2]string{
		MatchupTeamLabel(symGet(qf, 2).AwayShort, symGet(qf, 2).AwayTeam, symGet(qf, 2).TBDAway),
		MatchupTeamLabel(symGet(qf, 3).AwayShort, symGet(qf, 3).AwayTeam, symGet(qf, 3).TBDAway),
	}

	lQF0w := nextRoundTeamName(symGet(qf, 0))
	lQF1w := nextRoundTeamName(symGet(qf, 1))
	rQF2w := nextRoundTeamName(symGet(qf, 2))
	rQF3w := nextRoundTeamName(symGet(qf, 3))
	lSFw := nextRoundTeamName(symGet(sf, 0))
	rSFw := nextRoundTeamName(symGet(sf, 1))

	c := ConnectorStyle.Render
	t := MatchLineStyle.Render
	w := WinnerStyle.Render

	qfSp := strings.Repeat(" ", 9)
	sfSp := strings.Repeat(" ", 20)
	sfMid := strings.Repeat(" ", 11)
	rSp := strings.Repeat(" ", 9)

	ll := make([]string, 7)
	ll[0] = t(lH[0]) + c(" ─┐")
	ll[1] = qfSp + c("├─ ") + w(lQF0w) + c(" ─┐")
	ll[2] = t(lA[0]) + c(" ─┘") + sfMid + c("│")
	ll[3] = sfSp + c("├─ ") + w(lSFw)
	ll[4] = t(lH[1]) + c(" ─┐") + sfMid + c("│")
	ll[5] = qfSp + c("├─ ") + w(lQF1w) + c(" ─┘")
	ll[6] = t(lA[1]) + c(" ─┘")

	rr := make([]string, 7)
	rr[0] = " " + rSp + c("┌─ ") + t(rH[0])
	rr[1] = " " + w(rQF2w) + c(" ─┤")
	rr[2] = c("│") + rSp + c("└─ ") + t(rA[0])
	rr[3] = c("├─ ") + w(rSFw)
	rr[4] = c("│") + rSp + c("┌─ ") + t(rH[1])
	rr[5] = " " + w(rQF3w) + c(" ─┘")
	rr[6] = " " + rSp + c("└─ ") + t(rA[1])

	return symJoin(ll, rr, 7, symFinalLabel(fin, wcData))
}

// symFinalLabel returns the center line label for the Final matchup.
func symFinalLabel(fin []api.WCMatchup, wcData *api.WorldCupData) string {
	if wcData != nil && wcData.Champion != nil {
		return ChampionStyle.Render("🏆 " + TeamLabel(*wcData.Champion))
	}
	if len(fin) > 0 {
		mu := fin[0]
		home := MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
		away := MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway)
		if mu.HomeScore != nil && mu.AwayScore != nil {
			score := ScoreStyle.Render(fmt.Sprintf("%d–%d", *mu.HomeScore, *mu.AwayScore))
			if mu.IsPenalties {
				score += PenStyle.Render("p")
			}
			return WinnerStyle.Render(home) + " " + score + " " + WinnerStyle.Render(away)
		}
		if home != "TBD" || away != "TBD" {
			return MatchLineStyle.Render(home) + " " + ScoreStyle.Render("vs") + " " + MatchLineStyle.Render(away)
		}
	}
	return ScoreStyle.Render("── FINAL ──")
}

// symJoin combines left and right line slices with a fixed left-pad and a center label at midpoint.
func symJoin(ll, rr []string, n int, centerLabel string) string {
	const leftWidth = 32
	const centerWidth = 22
	mid := n / 2

	lines := make([]string, n)
	for i := 0; i < n; i++ {
		left := ""
		if i < len(ll) {
			left = ll[i]
		}
		right := ""
		if i < len(rr) {
			right = rr[i]
		}

		lw := lipgloss.Width(left)
		padL := ""
		if lw < leftWidth {
			padL = strings.Repeat(" ", leftWidth-lw)
		}

		var center string
		if i == mid {
			cw := lipgloss.Width(centerLabel)
			side := (centerWidth - cw) / 2
			if side < 0 {
				side = 0
			}
			center = strings.Repeat(" ", side) + centerLabel + strings.Repeat(" ", side)
		} else {
			center = strings.Repeat(" ", centerWidth)
		}

		lines[i] = left + padL + center + right
	}
	return strings.Join(lines, "\n")
}

// symGet safely indexes into a matchup slice, returning a TBD matchup if out of bounds.
func symGet(mus []api.WCMatchup, i int) api.WCMatchup {
	if i >= 0 && i < len(mus) {
		return mus[i]
	}
	return api.WCMatchup{TBDHome: true, TBDAway: true}
}

package worldcup

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// RenderSymmetricBracket renders the full knockout bracket in a single consolidated view.
// The layout automatically adapts:
//   - R32 present (2026): sym5Level — R32 compact feeders flank the R16→QF→SF→Final tree.
//   - R16 present (2022): sym4Level — R16→QF→SF→Final symmetric tree (unchanged style).
//   - Only QF present:    sym2Level — QF→SF→Final symmetric tree (unchanged style).
func RenderSymmetricBracket(width, height int, wcData *api.WorldCupData, banner string) string {
	if width <= 0 {
		width = 80
	}
	if wcData == nil {
		return LoadingStyle.Render("No bracket data")
	}

	header := design.RenderHeader(wcData.Name+" — Knockout Bracket", width-2)
	help := HelpStyle.Width(width).Render("u: upcoming  esc: back  q: quit")

	body := symBracketBody(wcData)

	parts := []string{}
	if banner != "" {
		parts = append(parts, banner)
	}
	parts = append(parts, header, "", body, "", help)
	return padToHeight(lipgloss.JoinVertical(lipgloss.Left, parts...), height)
}

// symBracketBody dispatches to the appropriate symmetric layout based on tournament depth.
func symBracketBody(wcData *api.WorldCupData) string {
	qf := symRound(wcData.KnockoutRounds, "1/4")
	sf := symRound(wcData.KnockoutRounds, "1/2")
	fin := symRound(wcData.KnockoutRounds, "final")
	if qf == nil || sf == nil || fin == nil {
		return LoadingStyle.Render("Bracket data not yet available")
	}
	r32 := symRound(wcData.KnockoutRounds, "1/16")
	r16 := symRound(wcData.KnockoutRounds, "1/8")
	if r32 != nil && r16 != nil {
		return sym5Level(r32.Matchups, r16.Matchups, qf.Matchups, sf.Matchups, fin.Matchups, wcData)
	}
	if r16 != nil {
		return sym4Level(r16.Matchups, qf.Matchups, sf.Matchups, fin.Matchups, wcData)
	}
	return sym2Level(qf.Matchups, sf.Matchups, fin.Matchups, wcData)
}

func symRound(rounds []api.WCKnockoutRound, stage string) *api.WCKnockoutRound {
	for i := range rounds {
		if rounds[i].Stage == stage {
			return &rounds[i]
		}
	}
	return nil
}

// ─── R32 → R16 → QF → SF → Final (2026 format) ───────────────────────────────

// sym5Level renders the full 2026 bracket in a single 15-line symmetric tree.
//
// Left half (outer → inner):
//
//	R32[0] compact ──► R16[0] home ─┐
//	                    [indent] ├─ QF[0] top ─┐
//	R32[1] compact ──► R16[0] away ─┘    sfMid │
//	                    [sfSp] ├─ SF[0] top → Final
//	R32[2] compact ──► R16[1] home ─┐    sfMid │
//	... (mirrored bottom half)
//
// Right half is the mirror: SF ← QF ← R16 ← R32 compact.
func sym5Level(r32, r16, qf, sf, fin []api.WCMatchup, wcData *api.WorldCupData) string {
	const r32ColW = 20 // fixed visual width for the R32 compact column (padded to align)
	const r32ArmW = 5  // " ──► "

	// r32SlotL renders one R32 match padded to r32ColW, then the arm connector.
	r32SlotL := func(i int) string {
		comp := symCompact(symGet(r32, i))
		w := lipgloss.Width(comp)
		if w < r32ColW {
			comp += strings.Repeat(" ", r32ColW-w)
		}
		return comp + ConnectorStyle.Render(" ──► ")
	}
	r32BlankL := strings.Repeat(" ", r32ColW+r32ArmW)

	// r32SlotR appends an R32 match after the R16 label on the right half.
	r32SlotR := func(i int) string {
		return "  " + symCompact(symGet(r32, i))
	}

	// r16Label derives the R16 slot label from the confirmed R32 match winner.
	// FotMob pre-seeds R16 team names before R32 matches are played (TBDHome=false
	// with actual team data). Reading R16 directly would show unearned teams.
	// We always derive from the R32 winner to stay consistent with match results.
	r16Label := func(r32Idx int) string {
		mu := symGet(r32, r32Idx)
		if mu.WinnerID != nil {
			return nextRoundTeamName(mu)
		}
		return padToLabelWidth("   ?")
	}

	// R16 team labels (left: 0-3, right: 4-7) — derived from R32 winners.
	// Left:  R16[i] home ← R32[i*2] winner, away ← R32[i*2+1] winner.
	// Right: R16[4+i] home ← R32[8+i*2] winner, away ← R32[8+i*2+1] winner.
	var lH, lA, rH, rA [4]string
	for i := range lH {
		lH[i] = MatchLineStyle.Render(r16Label(i * 2))
		lA[i] = MatchLineStyle.Render(r16Label(i*2 + 1))
	}
	for i := range rH {
		rH[i] = MatchLineStyle.Render(r16Label(8 + i*2))
		rA[i] = MatchLineStyle.Render(r16Label(8 + i*2 + 1))
	}

	// Winners advancing from R16 into QF slots
	lQF0h := nextRoundTeamName(symGet(r16, 0))
	lQF0a := nextRoundTeamName(symGet(r16, 1))
	lQF1h := nextRoundTeamName(symGet(r16, 2))
	lQF1a := nextRoundTeamName(symGet(r16, 3))
	rQF2h := nextRoundTeamName(symGet(r16, 4))
	rQF2a := nextRoundTeamName(symGet(r16, 5))
	rQF3h := nextRoundTeamName(symGet(r16, 6))
	rQF3a := nextRoundTeamName(symGet(r16, 7))

	// Winners advancing from QF into SF slots
	lSFh := nextRoundTeamName(symGet(qf, 0))
	lSFa := nextRoundTeamName(symGet(qf, 1))
	rSFh := nextRoundTeamName(symGet(qf, 2))
	rSFa := nextRoundTeamName(symGet(qf, 3))

	c := ConnectorStyle.Render
	qfSp := strings.Repeat(" ", 9)
	sfSp := strings.Repeat(" ", 20)
	sfMid := strings.Repeat(" ", 11)

	// Left block: 15 lines, R32 compact prepended on even (team) lines.
	ll := make([]string, 15)
	ll[0] = r32SlotL(0) + lH[0] + c(" ─┐")
	ll[1] = r32BlankL + qfSp + c("├─ ") + symTeamRender(symGet(qf, 0), true, WinnerStyle)(lQF0h) + c(" ─┐")
	ll[2] = r32SlotL(1) + lA[0] + c(" ─┘") + sfMid + c("│")
	ll[3] = r32BlankL + sfSp + c("├─ ") + symTeamRender(symGet(sf, 0), true, WinnerStyle)(lSFh)
	ll[4] = r32SlotL(2) + lH[1] + c(" ─┐") + sfMid + c("│")
	ll[5] = r32BlankL + qfSp + c("├─ ") + symTeamRender(symGet(qf, 0), false, WinnerStyle)(lQF0a) + c(" ─┘")
	ll[6] = r32SlotL(3) + lA[1] + c(" ─┘")
	ll[7] = r32BlankL + sfSp + c("│")
	ll[8] = r32SlotL(4) + lH[2] + c(" ─┐") + sfMid + c("│")
	ll[9] = r32BlankL + qfSp + c("├─ ") + symTeamRender(symGet(qf, 1), true, WinnerStyle)(lQF1h) + c(" ─┐")
	ll[10] = r32SlotL(5) + lA[2] + c(" ─┘") + sfMid + c("│")
	ll[11] = r32BlankL + sfSp + c("├─ ") + symTeamRender(symGet(sf, 0), false, WinnerStyle)(lSFa)
	ll[12] = r32SlotL(6) + lH[3] + c(" ─┐") + sfMid + c("│")
	ll[13] = r32BlankL + qfSp + c("├─ ") + symTeamRender(symGet(qf, 1), false, WinnerStyle)(lQF1a) + c(" ─┘")
	ll[14] = r32SlotL(7) + lA[3] + c(" ─┘")

	// Right block: 15 lines, mirror of left. R32 compact appended on team lines.
	// Line layout follows symRightQFBlock structure but inlined to allow R32 append.
	rSp := strings.Repeat(" ", 10)
	rr := make([]string, 15)
	// Top right: R16[4,5] → QF[2] → SF right top
	rr[0] = " " + rSp + c("┌─ ") + rH[0] + r32SlotR(8)
	rr[1] = c("┌─ ") + symTeamRender(symGet(qf, 2), true, WinnerStyle)(rQF2h) + c(" ─┤")
	rr[2] = c("│") + rSp + c("└─ ") + rA[0] + r32SlotR(9)
	rr[3] = c("┤")
	rr[4] = c("│") + rSp + c("┌─ ") + rH[1] + r32SlotR(10)
	rr[5] = c("└─ ") + symTeamRender(symGet(qf, 2), false, WinnerStyle)(rQF2a) + c(" ─┘")
	rr[6] = " " + rSp + c("└─ ") + rA[1] + r32SlotR(11)
	rr[7] = c("│")
	// Bottom right: R16[6,7] → QF[3] → SF right bottom
	rr[8] = " " + rSp + c("┌─ ") + rH[2] + r32SlotR(12)
	rr[9] = c("┌─ ") + symTeamRender(symGet(qf, 3), true, WinnerStyle)(rQF3h) + c(" ─┤")
	rr[10] = c("│") + rSp + c("└─ ") + rA[2] + r32SlotR(13)
	rr[11] = c("┤")
	rr[12] = c("│") + rSp + c("┌─ ") + rH[3] + r32SlotR(14)
	rr[13] = c("└─ ") + symTeamRender(symGet(qf, 3), false, WinnerStyle)(rQF3a) + c(" ─┘")
	rr[14] = " " + rSp + c("└─ ") + rA[3] + r32SlotR(15)

	finalLabel := symFinalLabel(fin, wcData)
	centers := symBuildCenters(15, finalLabel, map[int]string{
		3:  symTeamRender(symGet(sf, 1), true, WinnerStyle)(rSFh),
		11: symTeamRender(symGet(sf, 1), false, WinnerStyle)(rSFa),
	})
	return symJoin(ll, rr, 15, centers)
}

// ─── R16 → QF → SF → Final (2022 format) ─────────────────────────────────────

// sym4Level builds a 15-line symmetric bracket for R16 → QF → SF → Final.
//
// Left side column positions (visual chars):
//
//	col0: R16 team label (0-8, 9 chars: label6 + " ─┐"3)
//	col1: QF connector (9-20: sp9 + "├─ " + label6 + " ─┐"3 = 21 chars)
//	sfCol: SF connector │ at position 20
func sym4Level(r16, qf, sf, fin []api.WCMatchup, wcData *api.WorldCupData) string {
	var lH, lA, rH, rA [4]string
	for i := range lH {
		mu := symGet(r16, i)
		lH[i] = symTeamRender(mu, true, MatchLineStyle)(MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome))
		lA[i] = symTeamRender(mu, false, MatchLineStyle)(MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway))
	}
	for i := range rH {
		mu := symGet(r16, 4+i)
		rH[i] = symTeamRender(mu, true, MatchLineStyle)(MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome))
		rA[i] = symTeamRender(mu, false, MatchLineStyle)(MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway))
	}

	lQF0h := nextRoundTeamName(symGet(r16, 0))
	lQF0a := nextRoundTeamName(symGet(r16, 1))
	lQF1h := nextRoundTeamName(symGet(r16, 2))
	lQF1a := nextRoundTeamName(symGet(r16, 3))
	rQF2h := nextRoundTeamName(symGet(r16, 4))
	rQF2a := nextRoundTeamName(symGet(r16, 5))
	rQF3h := nextRoundTeamName(symGet(r16, 6))
	rQF3a := nextRoundTeamName(symGet(r16, 7))

	lSFh := nextRoundTeamName(symGet(qf, 0))
	lSFa := nextRoundTeamName(symGet(qf, 1))
	rSFh := nextRoundTeamName(symGet(qf, 2))
	rSFa := nextRoundTeamName(symGet(qf, 3))

	c := ConnectorStyle.Render
	qfSp := strings.Repeat(" ", 9)
	sfSp := strings.Repeat(" ", 20)
	sfMid := strings.Repeat(" ", 11)

	ll := make([]string, 15)
	ll[0] = lH[0] + c(" ─┐")
	ll[1] = qfSp + c("├─ ") + symTeamRender(symGet(qf, 0), true, WinnerStyle)(lQF0h) + c(" ─┐")
	ll[2] = lA[0] + c(" ─┘") + sfMid + c("│")
	ll[3] = sfSp + c("├─ ") + symTeamRender(symGet(sf, 0), true, WinnerStyle)(lSFh)
	ll[4] = lH[1] + c(" ─┐") + sfMid + c("│")
	ll[5] = qfSp + c("├─ ") + symTeamRender(symGet(qf, 0), false, WinnerStyle)(lQF0a) + c(" ─┘")
	ll[6] = lA[1] + c(" ─┘")
	ll[7] = sfSp + c("│")
	ll[8] = lH[2] + c(" ─┐") + sfMid + c("│")
	ll[9] = qfSp + c("├─ ") + symTeamRender(symGet(qf, 1), true, WinnerStyle)(lQF1h) + c(" ─┐")
	ll[10] = lA[2] + c(" ─┘") + sfMid + c("│")
	ll[11] = sfSp + c("├─ ") + symTeamRender(symGet(sf, 0), false, WinnerStyle)(lSFa)
	ll[12] = lH[3] + c(" ─┐") + sfMid + c("│")
	ll[13] = qfSp + c("├─ ") + symTeamRender(symGet(qf, 1), false, WinnerStyle)(lQF1a) + c(" ─┘")
	ll[14] = lA[3] + c(" ─┘")

	rSp := strings.Repeat(" ", 10)
	rr := make([]string, 15)
	copy(rr, symRightQFBlock(
		rH[0], rA[0], symTeamRender(symGet(qf, 2), true, WinnerStyle)(rQF2h),
		rH[1], rA[1], symTeamRender(symGet(qf, 2), false, WinnerStyle)(rQF2a),
		rSp,
	))
	rr[7] = c("│")
	copy(rr[8:], symRightQFBlock(
		rH[2], rA[2], symTeamRender(symGet(qf, 3), true, WinnerStyle)(rQF3h),
		rH[3], rA[3], symTeamRender(symGet(qf, 3), false, WinnerStyle)(rQF3a),
		rSp,
	))

	finalLabel := symFinalLabel(fin, wcData)
	centers := symBuildCenters(15, finalLabel, map[int]string{
		3:  symTeamRender(symGet(sf, 1), true, WinnerStyle)(rSFh),
		11: symTeamRender(symGet(sf, 1), false, WinnerStyle)(rSFa),
	})
	return symJoin(ll, rr, 15, centers)
}

// sym2Level builds a 7-line symmetric bracket for QF → SF → Final (2022 format).
func sym2Level(qf, sf, fin []api.WCMatchup, wcData *api.WorldCupData) string {
	var lH, lA, rH, rA [2]string
	for i := range lH {
		mu := symGet(qf, i)
		lH[i] = symTeamRender(mu, true, MatchLineStyle)(MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome))
		lA[i] = symTeamRender(mu, false, MatchLineStyle)(MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway))
	}
	for i := range rH {
		mu := symGet(qf, 2+i)
		rH[i] = symTeamRender(mu, true, MatchLineStyle)(MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome))
		rA[i] = symTeamRender(mu, false, MatchLineStyle)(MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway))
	}

	lQF0w := nextRoundTeamName(symGet(qf, 0))
	lQF1w := nextRoundTeamName(symGet(qf, 1))
	rQF2w := nextRoundTeamName(symGet(qf, 2))
	rQF3w := nextRoundTeamName(symGet(qf, 3))
	lSFw := nextRoundTeamName(symGet(sf, 0))
	rSFw := nextRoundTeamName(symGet(sf, 1))

	c := ConnectorStyle.Render
	qfSp := strings.Repeat(" ", 9)
	sfSp := strings.Repeat(" ", 20)
	sfMid := strings.Repeat(" ", 11)
	rSp := strings.Repeat(" ", 10)

	ll := make([]string, 7)
	ll[0] = lH[0] + c(" ─┐")
	ll[1] = qfSp + c("├─ ") + symTeamRender(symGet(sf, 0), true, WinnerStyle)(lQF0w) + c(" ─┐")
	ll[2] = lA[0] + c(" ─┘") + sfMid + c("│")
	ll[3] = sfSp + c("├─ ") + WinnerStyle.Render(lSFw)
	ll[4] = lH[1] + c(" ─┐") + sfMid + c("│")
	ll[5] = qfSp + c("├─ ") + symTeamRender(symGet(sf, 0), false, WinnerStyle)(lQF1w) + c(" ─┘")
	ll[6] = lA[1] + c(" ─┘")

	rr := symRightQFBlock(
		rH[0], rA[0], symTeamRender(symGet(sf, 1), true, WinnerStyle)(rQF2w),
		rH[1], rA[1], symTeamRender(symGet(sf, 1), false, WinnerStyle)(rQF3w),
		rSp,
	)

	finalLabel := symFinalLabel(fin, wcData)
	centers := symBuildCenters(7, finalLabel, map[int]string{3: WinnerStyle.Render(rSFw)})
	return symJoin(ll, rr, 7, centers)
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

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
		if !mu.TBDHome || !mu.TBDAway {
			return MatchLineStyle.Render(home) + " " + ScoreStyle.Render("vs") + " " + MatchLineStyle.Render(away)
		}
	}
	return ScoreStyle.Render("── FINAL ──")
}

// symBuildCenters returns n center strings of width centerWidth.
// midLabel is centered at the midpoint row.
// rightSF maps row → styled label for right-side SF winners: each label is right-aligned
// in center with " ─" appended (the ─ arm leads into the rr[row]="┤" connector).
// When a row is both midpoint AND in rightSF (7-line bracket), both fit in the 22 chars.
func symBuildCenters(n int, midLabel string, rightSF map[int]string) []string {
	const centerWidth = 22
	mid := n / 2
	centers := make([]string, n)
	for i := range centers {
		rsfLabel, hasSF := rightSF[i]
		isMid := i == mid
		sfArm := ConnectorStyle.Render(" ─")
		switch {
		case hasSF && isMid:
			sfPart := rsfLabel + sfArm
			sfW := lipgloss.Width(sfPart)
			rem := centerWidth - sfW
			cw := lipgloss.Width(midLabel)
			lp := (rem - cw) / 2
			if lp < 0 {
				lp = 0
			}
			rp := rem - cw - lp
			if rp < 0 {
				rp = 0
			}
			centers[i] = strings.Repeat(" ", lp) + midLabel + strings.Repeat(" ", rp) + sfPart
		case hasSF:
			sfPart := rsfLabel + sfArm
			sfW := lipgloss.Width(sfPart)
			pad := centerWidth - sfW
			if pad < 0 {
				pad = 0
			}
			centers[i] = strings.Repeat(" ", pad) + sfPart
		case isMid:
			cw := lipgloss.Width(midLabel)
			side := (centerWidth - cw) / 2
			if side < 0 {
				side = 0
			}
			rside := centerWidth - cw - side
			if rside < 0 {
				rside = 0
			}
			centers[i] = strings.Repeat(" ", side) + midLabel + strings.Repeat(" ", rside)
		default:
			centers[i] = strings.Repeat(" ", centerWidth)
		}
	}
	return centers
}

// symJoin combines left and right line slices using pre-computed per-row center strings.
// leftWidth is computed dynamically from the widest ll line so sym5Level's extended
// left block is handled without a hardcoded constant.
func symJoin(ll, rr []string, n int, centers []string) string {
	leftWidth := 0
	for _, l := range ll {
		if w := lipgloss.Width(l); w > leftWidth {
			leftWidth = w
		}
	}
	leftWidth++ // one cell of breathing room

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
		center := ""
		if i < len(centers) {
			center = centers[i]
		}
		lines[i] = left + padL + center + right
	}
	return strings.Join(lines, "\n")
}

// symRightQFBlock builds the 7-line right-side bracket for a pair of QF matches.
func symRightQFBlock(topH, topA, winTop, botH, botA, winBot, rSp string) []string {
	c := ConnectorStyle.Render
	return []string{
		" " + rSp + c("┌─ ") + topH,
		c("┌─ ") + winTop + c(" ─┤"),
		c("│") + rSp + c("└─ ") + topA,
		c("┤"),
		c("│") + rSp + c("┌─ ") + botH,
		c("└─ ") + winBot + c(" ─┘"),
		" " + rSp + c("└─ ") + botA,
	}
}

// symCompact renders a matchup as a single "home score away" line for the R32 column.
func symCompact(mu api.WCMatchup) string {
	home := MatchupTeamLabel(mu.HomeShort, mu.HomeTeam, mu.TBDHome)
	away := MatchupTeamLabel(mu.AwayShort, mu.AwayTeam, mu.TBDAway)
	hs, as_ := MatchLineStyle, MatchLineStyle
	if mu.WinnerID != nil {
		if *mu.WinnerID == mu.HomeTeamID {
			hs, as_ = WinnerStyle, EliminatedStyle
		} else {
			hs, as_ = EliminatedStyle, WinnerStyle
		}
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

// symGet safely indexes into a matchup slice, returning a TBD matchup if out of bounds.
func symGet(mus []api.WCMatchup, i int) api.WCMatchup {
	if i >= 0 && i < len(mus) {
		return mus[i]
	}
	return api.WCMatchup{TBDHome: true, TBDAway: true}
}

// symTeamRender returns a render func applying baseStyle normally, but switching
// to EliminatedStyle when the match is settled and this team lost.
func symTeamRender(mu api.WCMatchup, isHome bool, baseStyle lipgloss.Style) func(string) string {
	style := baseStyle
	if mu.WinnerID != nil {
		homeWon := *mu.WinnerID == mu.HomeTeamID
		if homeWon != isHome {
			style = EliminatedStyle
		}
	}
	return func(s string) string { return style.Render(s) }
}

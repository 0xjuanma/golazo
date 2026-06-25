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
	w := WinnerStyle.Render
	indent := strings.Repeat(" ", labelTargetWidth+3)
	gap := "     "

	l0 := symTeamRender(lmu, true, MatchLineStyle)(lh) + c(" ─┐")
	l1 := indent + c("├─ ") + w(lw)
	l2 := symTeamRender(lmu, false, MatchLineStyle)(la) + c(" ─┘")
	// Offset r0/r2 so ┌/└ align with ┤ in r1.
	// r1 = w(rw)(6) + " ─┤"(3) = 9 chars → ┤ at col 8; match by padding r0/r2 by same width.
	rwOff := strings.Repeat(" ", lipgloss.Width(w(rw))+2)
	r0 := rwOff + c("┌─ ") + symTeamRender(rmu, true, MatchLineStyle)(rh)
	r1 := w(rw) + c(" ─┤")
	r2 := rwOff + c("└─ ") + symTeamRender(rmu, false, MatchLineStyle)(ra)

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

	// sfCol=20: │ at visual position 20 in left block.
	// After label(6)+" ─┐"(3)=9 chars: need 11 more spaces to reach col 20.
	qfSp := strings.Repeat(" ", 9)   // indent before QF ├─
	sfSp := strings.Repeat(" ", 20)  // indent before SF ├─
	sfMid := strings.Repeat(" ", 11) // filler between R16 connector end (col9) and SF │ (col20)

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

	// Right side: SF column at col 0, QF bracket ┌/┤/└ at col 11.
	// "┌─ "(3) + label(6) + " ─┤"(3) = 12 chars → ┤ at col 11.
	// rSp = 10: │(col 0) + 10sp + ┌─/└─ lands at col 11.
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
		sfArm := ConnectorStyle.Render(" ─") // 2-char arm connecting center to rr[row]="┤"
		switch {
		case hasSF && isMid:
			// Midpoint row doubles as SF row (7-line): fit both labels in 22 chars.
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
func symJoin(ll, rr []string, n int, centers []string) string {
	const leftWidth = 32
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
// topH/topA and botH/botA are pre-styled team labels; winTop/winBot are pre-styled
// winner labels. rSp is the spacing between the SF column (col 0) and the QF arm.
// The returned lines slot directly into rr[0..6] or rr[8..14].
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

// symGet safely indexes into a matchup slice, returning a TBD matchup if out of bounds.
func symGet(mus []api.WCMatchup, i int) api.WCMatchup {
	if i >= 0 && i < len(mus) {
		return mus[i]
	}
	return api.WCMatchup{TBDHome: true, TBDAway: true}
}

// symTeamRender returns a render func that applies baseStyle normally, but
// switches to EliminatedStyle when the match is settled and this team lost.
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

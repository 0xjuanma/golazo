package worldcup

import "github.com/0xjuanma/golazo/internal/api"

// Phase constants used in the phase banner.
const (
	PhaseGroupStage    = "Group Stage"
	PhaseRoundOf32     = "Round of 32"
	PhaseRoundOf16     = "Round of 16"
	PhaseQuarterFinals = "Quarter-finals"
	PhaseSemiFinals    = "Semi-finals"
	PhaseFinal         = "Final"
	PhaseCompleted     = "Completed"
)

// DerivePhase returns the human-readable current tournament phase.
// Logic: scan knockout rounds from earliest to latest; the first round
// with at least one unresolved matchup (WinnerID == nil) is the active phase.
// If all knockout matchups are resolved, the tournament is complete.
// If no knockout rounds exist, we're still in the group stage.
func DerivePhase(wcData *api.WorldCupData) string {
	if wcData == nil {
		return ""
	}
	if len(wcData.KnockoutRounds) == 0 {
		return PhaseGroupStage
	}

	stageMap := map[string]string{
		"1/32":  PhaseRoundOf32,
		"1/16":  PhaseRoundOf16,
		"1/8":   PhaseQuarterFinals,
		"1/4":   PhaseSemiFinals,
		"1/2":   PhaseFinal,
		"final": PhaseFinal,
	}

	for _, round := range wcData.KnockoutRounds {
		for _, mu := range round.Matchups {
			if mu.WinnerID == nil {
				if label, ok := stageMap[round.Stage]; ok {
					return label
				}
				return round.Label
			}
		}
	}

	// All resolved
	return PhaseCompleted
}

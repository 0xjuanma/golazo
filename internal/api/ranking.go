package api

// FIFARankingPeriod identifies one official FIFA ranking release.
type FIFARankingPeriod struct {
	ID   string
	Name string
}

// FIFARankingEntry is one national team in an official FIFA ranking.
type FIFARankingEntry struct {
	TeamID         int
	Name           string
	Rank           int
	TotalPoints    float64
	PreviousPoints float64
	PointsDiff     float64
	GainedRank     bool
	LostRank       bool
}

// FIFARanking contains one released men's or women's FIFA ranking.
type FIFARanking struct {
	Gender string
	Period FIFARankingPeriod
	Teams  []FIFARankingEntry
}

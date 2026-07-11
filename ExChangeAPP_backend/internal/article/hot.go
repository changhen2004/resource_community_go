package article

import "time"

const (
	hotScoreView     = 1
	hotScoreLike     = 8
	hotScoreComment  = 12
	hotScoreFavorite = 10
)

func initialHotScore(createdAt time.Time) float64 {
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	// Keep a modest freshness bias without letting publish time fully dominate engagement.
	return 50 + float64(createdAt.Unix())/86400
}

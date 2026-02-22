package analytics

// interestPoints maps event names to the base points awarded to the user's
// interest profile. Points are distributed proportionally across the manga's
// tags; author and category each receive the full base points.
var interestPoints = map[string]float64{
	"manga_view":       1,
	"chapter_open":     3,
	"chapter_complete": 5,
	"comment_post":     4,
	"bookmark_add":     8,
	"bookmark_remove":  -3,
}

// trendingPoints maps event names to the score increment applied to the
// global trending sorted set.
var trendingPoints = map[string]float64{
	"manga_view":       1,
	"chapter_open":     3,
	"chapter_complete": 5,
}

const (
	// interestDecay is the multiplier applied to the existing interest score
	// before adding new points: new = old * decay + points.
	// At 0.98, a score decays to ~1% of its original value after ~230 interactions
	// from other events, naturally fading old interests.
	interestDecay = 0.98

	// trendingDecay is applied to the entire trending set once per hour.
	// At 0.9, a score decays to ~1% after ~44 hours with no new activity.
	trendingDecay = 0.9
)

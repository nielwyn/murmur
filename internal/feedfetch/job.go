package feedfetch

import (
	"time"

	"github.com/nielwyn/murmur/internal/database"
)

// FetchJob is one unit of work: fetch a single feed and store any new posts.
type FetchJob struct {
	Feed database.Feed
}

type FetchResult struct {
	Feed     database.Feed
	NewPosts int
	Duration time.Duration
	Err      error
}

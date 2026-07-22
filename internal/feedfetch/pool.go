package feedfetch

import (
	"context"
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	"github.com/nielwyn/murmur/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/microcosm-cc/bluemonday"
)

var descriptionPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AddTargetBlankToFullyQualifiedLinks(true)
	return p
}()

// worker pulls jobs until the channel is closed, sending one FetchResult per
// job to results. Running several of these concurrently is the fan-out half
// of the pipeline.
func (s *Scheduler) worker(ctx context.Context, jobs <-chan FetchJob, results chan<- FetchResult) {
	for job := range jobs {
		results <- s.fetchOne(ctx, job.Feed)
	}
}

// fetchOne fetches a single feed and stores any new posts. fetchTimeout
// bounds the HTTP request so one slow/unresponsive feed can't stall a
// worker (and thus the whole batch) indefinitely.
func (s *Scheduler) fetchOne(ctx context.Context, feed database.Feed) FetchResult {
	start := time.Now()

	if _, err := s.db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return FetchResult{Feed: feed, Duration: time.Since(start), Err: fmt.Errorf("marking feed fetched: %w", err)}
	}

	fetchCtx, cancel := context.WithTimeout(ctx, s.cfg.FetchTimeout)
	defer cancel()

	fetchedFeed, err := Fetch(fetchCtx, feed.Link)
	if err != nil {
		return FetchResult{Feed: feed, Duration: time.Since(start), Err: fmt.Errorf("fetching feed: %w", err)}
	}

	newPosts := 0
	for _, item := range fetchedFeed.Items {
		if item.Link == "" {
			continue
		}
		var publishedAt = pgtype.Timestamp{Time: *item.PublishedParsed, Valid: true}
		description := strings.TrimSpace(descriptionPolicy.Sanitize(html.UnescapeString(item.Description)))
		rows, err := s.db.CreatePost(ctx, database.CreatePostParams{
			Title:       html.UnescapeString(item.Title),
			Link:        item.Link,
			Description: &description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			log.Printf("feedfetch: saving post %q: %v", item.Title, err)
			continue
		}
		newPosts += int(rows)
	}

	return FetchResult{Feed: feed, NewPosts: newPosts, Duration: time.Since(start)}
}

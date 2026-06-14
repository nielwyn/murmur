package feedfetch

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"murmur/internal/database"
	"murmur/internal/rssfeed"

	"github.com/jackc/pgx/v5/pgtype"
)

// worker pulls jobs until the channel is closed, sending one FetchResult per
// job to results. Running several of these concurrently is the fan-out half
// of the pipeline.
func worker(ctx context.Context, db *database.Queries, fetchTimeout time.Duration, jobs <-chan FetchJob, results chan<- FetchResult) {
	for job := range jobs {
		results <- fetchOne(ctx, db, fetchTimeout, job.Feed)
	}
}

// fetchOne fetches a single feed and stores any new posts. fetchTimeout
// bounds the HTTP request so one slow/unresponsive feed can't stall a
// worker (and thus the whole batch) indefinitely.
func fetchOne(ctx context.Context, db *database.Queries, fetchTimeout time.Duration, feed database.Feed) FetchResult {
	start := time.Now()

	if _, err := db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return FetchResult{Feed: feed, Duration: time.Since(start), Err: fmt.Errorf("marking feed fetched: %w", err)}
	}

	fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	rss, err := rssfeed.Fetch(fetchCtx, feed.Url)
	if err != nil {
		return FetchResult{Feed: feed, Duration: time.Since(start), Err: fmt.Errorf("fetching feed: %w", err)}
	}

	newPosts := 0
	for _, item := range rss.Channel.Items {
		if item.Link == "" {
			continue
		}

		var publishedAt pgtype.Timestamp
		if t, err := rssfeed.ParseDate(item.PubDate); err == nil {
			publishedAt = pgtype.Timestamp{Time: t, Valid: true}
		}

		description := strings.TrimSpace(html.UnescapeString(item.Description))

		rows, err := db.CreatePost(ctx, database.CreatePostParams{
			Title:       html.UnescapeString(item.Title),
			Url:         item.Link,
			Description: &description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			continue
		}
		newPosts += int(rows)
	}

	return FetchResult{Feed: feed, NewPosts: newPosts, Duration: time.Since(start)}
}

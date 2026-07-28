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
	"github.com/mmcdole/gofeed"
)

var descriptionPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AddTargetBlankToFullyQualifiedLinks(true)
	return p
}()

// worker drains jobs until the channel closes, one FetchResult per job;
// run several for the fan-out half of the pipeline.
func (s *Scheduler) worker(ctx context.Context, jobs <-chan FetchJob, results chan<- FetchResult) {
	for job := range jobs {
		results <- s.fetchOne(ctx, job.Feed)
	}
}

// fetchTimeout bounds the request so one slow feed can't stall the worker.
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
		var publishedAt pgtype.Timestamp
		if item.PublishedParsed != nil {
			publishedAt = pgtype.Timestamp{Time: *item.PublishedParsed, Valid: true}
		}

		description := strings.TrimSpace(descriptionPolicy.Sanitize(html.UnescapeString(item.Description)))
		rows, err := s.db.CreatePost(ctx, database.CreatePostParams{
			Title:       html.UnescapeString(item.Title),
			Link:        item.Link,
			Description: &description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
			ImageUrl:    resolveImageURL(item),
		})
		if err != nil {
			log.Printf("feedfetch: saving post %q: %v", item.Title, err)
			continue
		}
		newPosts += int(rows)
	}

	return FetchResult{Feed: feed, NewPosts: newPosts, Duration: time.Since(start)}
}

// Falls back to media:thumbnail since gofeed's Image field misses it
// (e.g. BBC feeds only set media:thumbnail, not itunes:image/media:content).
func resolveImageURL(item *gofeed.Item) *string {
	if item.Image != nil && item.Image.URL != "" {
		url := item.Image.URL
		return &url
	}
	if thumbs := item.Extensions["media"]["thumbnail"]; len(thumbs) > 0 {
		if url := thumbs[0].Attrs["url"]; url != "" {
			return &url
		}
	}
	return nil
}

package main

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/nielwyn/murmur/internal/database"
	"github.com/nielwyn/murmur/internal/rssfeed"

	"github.com/jackc/pgx/v5/pgtype"
)

// handlerAgg runs a sequential feed-fetch loop on a ticker: one feed at a
// time, single goroutine. This is a placeholder for weeks 1-2 and will be
// replaced in week 5 by internal/feedfetch's worker-pool scheduler
// (fan-out/fan-in, per-fetch context timeouts, status map, rate limiter).
func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: agg <interval, e.g. 1m>")
	}
	interval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", cmd.args[0], err)
	}

	fmt.Printf("Collecting feeds every %s (Ctrl-C to stop)\n", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Println("error scraping feeds:", err)
		}
	}
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()

	feeds, err := s.db.GetFeedsDueForFetch(ctx, 10)
	if err != nil {
		return fmt.Errorf("getting feeds due for fetch: %w", err)
	}
	if len(feeds) == 0 {
		fmt.Println("no feeds due for fetching")
		return nil
	}

	for _, feed := range feeds {
		if err := scrapeFeed(ctx, s, feed); err != nil {
			fmt.Printf("error scraping feed %q: %v\n", feed.Name, err)
		}
	}
	return nil
}

func scrapeFeed(ctx context.Context, s *state, feed database.Feed) error {
	if _, err := s.db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return fmt.Errorf("marking feed fetched: %w", err)
	}

	rss, err := rssfeed.Fetch(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
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

		rows, err := s.db.CreatePost(ctx, database.CreatePostParams{
			Title:       html.UnescapeString(item.Title),
			Url:         item.Link,
			Description: &description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			fmt.Printf("  error saving post %q: %v\n", item.Title, err)
			continue
		}
		newPosts += int(rows)
	}

	fmt.Printf("* %s: %d new post(s)\n", feed.Name, newPosts)
	return nil
}

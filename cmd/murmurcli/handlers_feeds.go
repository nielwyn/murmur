package main

import (
	"context"
	"fmt"
)

func (s *state) handleListFeeds(cmd command) error {
	feeds, err := s.client.ListFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds yet. Add one with `addfeed <name> <url>`.")
		return nil
	}

	for _, f := range feeds {
		fmt.Printf("* %s (%s) - added by %s\n", f.Name, f.URL, f.CreatorName)
	}
	return nil
}

func (s *state) handleCreateFeed(cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	name, url := cmd.args[0], cmd.args[1]

	feed, err := s.client.CreateFeed(context.Background(), name, url)
	if err != nil {
		return fmt.Errorf("creating feed: %w", err)
	}

	if _, err := s.client.FollowFeed(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("following feed: %w", err)
	}

	fmt.Printf("Added feed %q (%s)\n", feed.Name, feed.URL)
	return nil
}

func (s *state) handleListFollowing(cmd command) error {
	follows, err := s.client.ListFollowing(context.Background())
	if err != nil {
		return fmt.Errorf("listing followed feeds: %w", err)
	}

	if len(follows) == 0 {
		fmt.Println("Not following any feeds yet.")
		return nil
	}

	fmt.Printf("%s follows:\n", s.cfg.CurrentUsername)
	for _, f := range follows {
		fmt.Printf("* %s (%s)\n", f.FeedName, f.FeedURL)
	}
	return nil
}

// findFeedByURL resolves a URL to a feed by listing feeds, since the API
// follows/unfollows by ID only.
func (s *state) findFeedByURL(ctx context.Context, url string) (feedResponse, error) {
	feeds, err := s.client.ListFeeds(ctx)
	if err != nil {
		return feedResponse{}, fmt.Errorf("looking up feed: %w", err)
	}
	for _, f := range feeds {
		if f.URL == url {
			return f, nil
		}
	}
	return feedResponse{}, fmt.Errorf("feed %q not found", url)
}

func (s *state) handleFollowFeed(cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}
	url := cmd.args[0]

	feed, err := s.findFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	if _, err := s.client.FollowFeed(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("following feed: %w", err)
	}

	fmt.Printf("%s now follows %q\n", s.cfg.CurrentUsername, feed.Name)
	return nil
}

func (s *state) handleUnfollowFeed(cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}
	url := cmd.args[0]

	feed, err := s.findFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	if err := s.client.UnfollowFeed(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("unfollowing feed: %w", err)
	}

	fmt.Printf("%s unfollowed %q\n", s.cfg.CurrentUsername, feed.Name)
	return nil
}

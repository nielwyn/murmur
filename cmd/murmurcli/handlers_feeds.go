package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/nielwyn/murmur/internal/database"
	"github.com/nielwyn/murmur/internal/service"
)

func (s *state) handleListFeeds(cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds yet. Add one with `addfeed <name> <url>`.")
		return nil
	}

	for _, f := range feeds {
		fmt.Printf("* %s (%s) - added by %s\n", f.Name, f.Url, f.CreatorName)
	}
	return nil
}

func (s *state) handleCreateFeed(cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	name, url := cmd.args[0], cmd.args[1]

	feed, err := s.svc.CreateFeed(context.Background(), user.ID, name, url)
	if err != nil {
		if errors.Is(err, service.ErrFeedExists) {
			return fmt.Errorf("a feed with this url already exists")
		}
		return fmt.Errorf("creating feed: %w", err)
	}

	if _, err := s.svc.FollowFeed(context.Background(), user.ID, feed.ID); err != nil {
		return fmt.Errorf("following feed: %w", err)
	}

	fmt.Printf("Added feed %q (%s)\n", feed.Name, feed.Url)
	return nil
}

func (s *state) handleListFollowing(cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("listing followed feeds: %w", err)
	}

	if len(follows) == 0 {
		fmt.Println("Not following any feeds yet.")
		return nil
	}

	fmt.Printf("%s follows:\n", user.Username)
	for _, f := range follows {
		fmt.Printf("* %s (%s)\n", f.FeedName, f.FeedUrl)
	}
	return nil
}

func (s *state) handleFollowFeed(cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("feed %q not found: %w", url, err)
	}

	if _, err := s.svc.FollowFeed(context.Background(), user.ID, feed.ID); err != nil {
		return fmt.Errorf("following feed: %w", err)
	}

	fmt.Printf("%s now follows %q\n", user.Username, feed.Name)
	return nil
}

func (s *state) handleUnfollowFeed(cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("feed %q not found: %w", url, err)
	}

	if err := s.svc.UnfollowFeed(context.Background(), user.ID, feed.ID); err != nil {
		if errors.Is(err, service.ErrNotFollowing) {
			return fmt.Errorf("you don't follow %q", feed.Name)
		}
		return fmt.Errorf("unfollowing feed: %w", err)
	}

	fmt.Printf("%s unfollowed %q\n", user.Username, feed.Name)
	return nil
}

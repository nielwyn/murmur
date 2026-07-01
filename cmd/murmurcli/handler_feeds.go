package main

import (
	"context"
	"fmt"

	"github.com/nielwyn/murmur/internal/database"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	name, url := cmd.args[0], cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		Name:   name,
		Url:    url,
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("creating feed: %w", err)
	}

	if _, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}); err != nil {
		return fmt.Errorf("following feed: %w", err)
	}

	fmt.Printf("Added feed %q (%s)\n", feed.Name, feed.Url)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("feed %q not found: %w", url, err)
	}

	if _, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}); err != nil {
		return fmt.Errorf("following feed: %w", err)
	}

	fmt.Printf("%s now follows %q\n", user.Name, feed.Name)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("feed %q not found: %w", url, err)
	}

	rows, err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("unfollowing feed: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("you don't follow %q", feed.Name)
	}

	fmt.Printf("%s unfollowed %q\n", user.Name, feed.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("listing followed feeds: %w", err)
	}

	if len(follows) == 0 {
		fmt.Println("Not following any feeds yet.")
		return nil
	}

	fmt.Printf("%s follows:\n", user.Name)
	for _, f := range follows {
		fmt.Printf("* %s (%s)\n", f.FeedName, f.FeedUrl)
	}
	return nil
}

package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nielwyn/murmur/internal/database"
)

func (s *state) handleListPosts(cmd command, user database.User) error {
	limit := 10
	if len(cmd.args) == 1 {
		n, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit %q: %w", cmd.args[0], err)
		}
		limit = n
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("listing posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts yet. Follow some feeds and let the fetcher run.")
		return nil
	}

	for _, p := range posts {
		published := "unknown date"
		if p.PublishedAt.Valid {
			published = p.PublishedAt.Time.Format("2006-01-02")
		}
		fmt.Printf("[%s] %s: %s\n  %s\n", published, p.FeedName, p.Title, p.Url)
	}
	return nil
}

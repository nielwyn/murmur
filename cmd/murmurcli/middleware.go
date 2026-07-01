package main

import (
	"context"
	"fmt"

	"github.com/nielwyn/murmur/internal/database"
)

// middlewareLoggedIn looks up the user named in the config file and passes
// it to the wrapped handler, so handlers don't each repeat the lookup.
func middlewareLoggedIn(handler func(*state, command, database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("not logged in: run `register` or `login` first")
		}

		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("could not find current user %q: %w", s.cfg.CurrentUserName, err)
		}

		return handler(s, cmd, user)
	}
}

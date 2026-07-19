package main

import (
	"context"
	"fmt"

	"github.com/nielwyn/murmur/internal/database"
)

// requireAuth resolves cfg.CurrentUsername to a database.User for the
// wrapped handler. This trusts the config file — there is no password or
// session check beyond `login` having written the name there.
func (s *state) requireAuth(next func(command, database.User) error) func(command) error {
	return func(cmd command) error {
		if s.cfg.CurrentUsername == "" {
			return fmt.Errorf("not logged in: run `register` or `login` first")
		}

		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUsername)
		if err != nil {
			return fmt.Errorf("could not find current user %q: %w", s.cfg.CurrentUsername, err)
		}

		return next(cmd, user)
	}
}

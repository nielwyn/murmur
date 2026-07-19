package main

import "fmt"

// requireAuth fails fast when there's no saved session token; the API
// still validates it on every request.
func (s *state) requireAuth(next func(command) error) func(command) error {
	return func(cmd command) error {
		if s.cfg.AuthToken == "" {
			return fmt.Errorf("not logged in: run `register` or `login` first")
		}
		return next(cmd)
	}
}

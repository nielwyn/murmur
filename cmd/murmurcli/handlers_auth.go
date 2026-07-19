package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/nielwyn/murmur/internal/service"
)

func (s *state) handleRegister(cmd command) error {
	if len(cmd.args) != 3 {
		return fmt.Errorf("usage: register <username> <email> <password>")
	}
	username, email, password := cmd.args[0], cmd.args[1], cmd.args[2]

	user, err := s.svc.Register(context.Background(), username, email, password)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	if err := s.cfg.SetUser(user.Username); err != nil {
		return fmt.Errorf("saving current user: %w", err)
	}

	fmt.Printf("User %q created and logged in\n", user.Username)
	return nil
}

func (s *state) handleLogin(cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: login <username> <password>")
	}
	username, password := cmd.args[0], cmd.args[1]

	user, err := s.svc.Login(context.Background(), username, password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return fmt.Errorf("username %q not found", username)
		}
		return fmt.Errorf("invalid password")
	}

	if err := s.cfg.SetUser(user.Username); err != nil {
		return fmt.Errorf("saving current user: %w", err)
	}

	fmt.Printf("Logged in as %q\n", user.Username)
	return nil
}

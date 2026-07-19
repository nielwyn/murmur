package main

import (
	"context"
	"fmt"
)

func (s *state) handleRegister(cmd command) error {
	if len(cmd.args) != 3 {
		return fmt.Errorf("usage: register <username> <email> <password>")
	}
	username, email, password := cmd.args[0], cmd.args[1], cmd.args[2]

	user, token, err := s.client.Register(context.Background(), username, email, password)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	if err := s.cfg.SetSession(user.Username, token); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	fmt.Printf("User %q created and logged in\n", user.Username)
	return nil
}

func (s *state) handleLogin(cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: login <username> <password>")
	}
	username, password := cmd.args[0], cmd.args[1]

	user, token, err := s.client.Login(context.Background(), username, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := s.cfg.SetSession(user.Username, token); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	fmt.Printf("Logged in as %q\n", user.Username)
	return nil
}

package main

import (
	"context"
	"fmt"

	"github.com/nielwyn/murmur/internal/database"

	"golang.org/x/crypto/bcrypt"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 3 {
		return fmt.Errorf("usage: register <name> <email> <password>")
	}
	name, email, password := cmd.args[0], cmd.args[1], cmd.args[2]

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		Name:           name,
		Email:          email,
		HashedPassword: string(hashed),
	})
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("saving current user: %w", err)
	}

	fmt.Printf("User %q created and logged in\n", user.Name)
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: login <name> <password>")
	}
	name, password := cmd.args[0], cmd.args[1]

	user, err := s.db.GetUserByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user %q not found: %w", name, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("saving current user: %w", err)
	}

	fmt.Printf("Logged in as %q\n", user.Name)
	return nil
}

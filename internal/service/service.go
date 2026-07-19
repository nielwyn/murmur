// Package service holds the business logic shared by the API server and the CLI
package service

import (
	"context"
	"errors"

	"github.com/nielwyn/murmur/internal/auth"
	"github.com/nielwyn/murmur/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

// Reports whether err is a Postgres unique-constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrFeedExists      = errors.New("feed already exists")
	ErrNotFollowing    = errors.New("not following feed")
)

type Service struct {
	db *database.Queries
}

func New(db *database.Queries) *Service {
	return &Service{db: db}
}

func (s *Service) Register(ctx context.Context, username, email, password string) (database.User, error) {
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return database.User{}, err
	}

	return s.db.CreateUser(ctx, database.CreateUserParams{
		Username:       username,
		Email:          email,
		HashedPassword: hashed,
	})
}

func (s *Service) Login(ctx context.Context, username, password string) (database.User, error) {
	user, err := s.db.GetUserByName(ctx, username)
	if err != nil {
		return database.User{}, ErrUserNotFound
	}

	if err := auth.CheckPassword(user.HashedPassword, password); err != nil {
		return database.User{}, ErrInvalidPassword
	}

	return user, nil
}

func (s *Service) CreateFeed(ctx context.Context, userID uuid.UUID, name, url string) (database.Feed, error) {
	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		Name:   name,
		Url:    url,
		UserID: userID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return database.Feed{}, ErrFeedExists
		}
		return database.Feed{}, err
	}

	return feed, nil
}

func (s *Service) FollowFeed(ctx context.Context, userID, feedID uuid.UUID) (database.FeedFollow, error) {
	return s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: userID,
		FeedID: feedID,
	})
}

func (s *Service) UnfollowFeed(ctx context.Context, userID, feedID uuid.UUID) error {
	rows, err := s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: userID,
		FeedID: feedID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFollowing
	}

	return nil
}

package api

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nielwyn/murmur/internal/database"
	"github.com/nielwyn/murmur/internal/feedfetch"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

const createFeedFetchTimeout = 10 * time.Second

const pgUniqueViolation = "23505"

// Reports whether err is a Postgres unique-constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

type feedResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	CreatorName string    `json:"creator_name,omitempty"`
}

type followResponse struct {
	FeedID   uuid.UUID `json:"feed_id"`
	FeedName string    `json:"feed_name,omitempty"`
	FeedURL  string    `json:"feed_url,omitempty"`
}

func (s *Server) handleListFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := s.db.GetFeeds(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not list feeds")
		return
	}

	resp := make([]feedResponse, len(feeds))
	for i, f := range feeds {
		resp[i] = feedResponse{ID: f.ID, Name: f.Name, URL: f.Url, CreatorName: f.CreatorName}
	}
	respondJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCreateFeed(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)

	var req struct {
		Url string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Url == "" {
		respondError(w, http.StatusBadRequest, "url is required")
		return
	}

	fetchCtx, cancel := context.WithTimeout(r.Context(), createFeedFetchTimeout)
	defer cancel()

	fetchedFeed, err := feedfetch.Fetch(fetchCtx, req.Url)
	if err != nil {
		respondError(w, http.StatusBadRequest, "could not fetch a feed at that url")
		return
	}

	feed, err := s.db.CreateFeed(r.Context(), database.CreateFeedParams{
		Name:   fetchedFeed.Title,
		Url:    cmp.Or(fetchedFeed.FeedLink, req.Url),
		UserID: user.ID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			respondError(w, http.StatusConflict, "a feed with this url already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "could not create feed")
		return
	}

	respondJSON(w, http.StatusCreated, feedResponse{ID: feed.ID, Name: feed.Name, URL: feed.Url})
}

func (s *Server) handleListFollowing(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)

	follows, err := s.db.GetFeedFollowsForUser(r.Context(), user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not list followed feeds")
		return
	}

	resp := make([]followResponse, len(follows))
	for i, f := range follows {
		resp[i] = followResponse{FeedID: f.FeedID, FeedName: f.FeedName, FeedURL: f.FeedUrl}
	}
	respondJSON(w, http.StatusOK, resp)
}

func (s *Server) handleFollowFeed(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)

	feedID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	follow, err := s.db.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	})
	if err != nil {
		respondError(w, http.StatusConflict, "could not follow feed")
		return
	}

	respondJSON(w, http.StatusCreated, followResponse{FeedID: follow.FeedID})
}

func (s *Server) handleUnfollowFeed(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)

	feedID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	rows, err := s.db.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not unfollow feed")
		return
	}
	if rows == 0 {
		respondError(w, http.StatusNotFound, "not following this feed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

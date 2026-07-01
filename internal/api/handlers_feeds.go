package api

import (
	"encoding/json"
	"net/http"

	"github.com/nielwyn/murmur/internal/database"

	"github.com/google/uuid"
)

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
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Url == "" {
		respondError(w, http.StatusBadRequest, "name and url are required")
		return
	}

	feed, err := s.db.CreateFeed(r.Context(), database.CreateFeedParams{
		Name:   req.Name,
		Url:    req.Url,
		UserID: user.ID,
	})
	if err != nil {
		respondError(w, http.StatusConflict, "could not create feed (url may already exist)")
		return
	}

	if _, err := s.db.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}); err != nil {
		respondError(w, http.StatusInternalServerError, "feed created but could not follow it")
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

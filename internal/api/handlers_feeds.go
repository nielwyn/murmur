package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nielwyn/murmur/internal/service"

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

	feed, err := s.svc.CreateFeed(r.Context(), user.ID, req.Name, req.Url)
	if err != nil {
		if errors.Is(err, service.ErrFeedExists) {
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

	follow, err := s.svc.FollowFeed(r.Context(), user.ID, feedID)
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

	if err := s.svc.UnfollowFeed(r.Context(), user.ID, feedID); err != nil {
		if errors.Is(err, service.ErrNotFollowing) {
			respondError(w, http.StatusNotFound, "not following this feed")
			return
		}
		respondError(w, http.StatusInternalServerError, "could not unfollow feed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

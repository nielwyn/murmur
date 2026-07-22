package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nielwyn/murmur/internal/database"
)

type postsResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Link        string     `json:"link"`
	Description *string    `json:"description,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	FeedID      uuid.UUID  `json:"feed_id"`
	FeedTitle   string     `json:"feed_title"`
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)

	limit := defaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			respondError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	posts, err := s.db.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not list followed feeds")
		return
	}
	resp := make([]postsResponse, len(posts))
	for i, p := range posts {
		var publishedAt *time.Time
		if p.PublishedAt.Valid {
			t := p.PublishedAt.Time
			publishedAt = &t
		}
		resp[i] = postsResponse{
			ID:          p.ID,
			Title:       p.Title,
			Link:        p.Link,
			Description: p.Description,
			PublishedAt: publishedAt,
			FeedID:      p.FeedID,
			FeedTitle:   p.FeedTitle,
		}
	}
	respondJSON(w, http.StatusOK, resp)
}

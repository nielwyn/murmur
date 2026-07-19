package api

import (
	"net/http"

	"github.com/nielwyn/murmur/internal/config"
	"github.com/nielwyn/murmur/internal/database"
	"github.com/nielwyn/murmur/internal/service"
)

type Server struct {
	db  *database.Queries
	cfg *config.Config
	svc *service.Service
}

func NewServer(db *database.Queries, cfg *config.Config) http.Handler {
	s := &Server{db: db, cfg: cfg, svc: service.New(db)}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", s.handleRegister)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)

	mux.HandleFunc("GET /api/me", s.requireAuth(s.handleMe))

	mux.HandleFunc("GET /api/feeds", s.requireAuth(s.handleListFeeds))
	mux.HandleFunc("POST /api/feeds", s.requireAuth(s.handleCreateFeed))
	mux.HandleFunc("GET /api/feeds/following", s.requireAuth(s.handleListFollowing))
	mux.HandleFunc("POST /api/feeds/{id}/follow", s.requireAuth(s.handleFollowFeed))
	mux.HandleFunc("DELETE /api/feeds/{id}/follow", s.requireAuth(s.handleUnfollowFeed))

	return recoverer(logger(mux))
}

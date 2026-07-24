package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nielwyn/murmur/internal/auth"
	"github.com/nielwyn/murmur/internal/database"

	"github.com/google/uuid"
)

const tokenExpiry = 7 * 24 * time.Hour

type userResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

func toUserResponse(u database.User) userResponse {
	return userResponse{ID: u.ID, Username: u.Username, Email: u.Email}
}

func (s *Server) issueSession(w http.ResponseWriter, user database.User) error {
	token, err := auth.MakeJWT(user.ID, s.cfg.JWTSecret, tokenExpiry)
	if err != nil {
		return err
	}
	auth.SetAuthCookie(w, token, tokenExpiry, s.cfg.Secure)
	return nil
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user, err := s.db.CreateUser(r.Context(), database.CreateUserParams{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: &hashed,
	})
	if err != nil {
		respondError(w, http.StatusConflict, "could not create user (username or email may already be taken)")
		return
	}

	if err := s.issueSession(w, user); err != nil {
		respondError(w, http.StatusInternalServerError, "could not create session")
		return
	}

	respondJSON(w, http.StatusCreated, toUserResponse(user))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.db.GetUserByName(r.Context(), req.UserName)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if user.HashedPassword == nil {
		respondError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if err := auth.CheckPassword(*user.HashedPassword, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if err := s.issueSession(w, user); err != nil {
		respondError(w, http.StatusInternalServerError, "could not create session")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearAuthCookie(w, s.cfg.Secure)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r)
	respondJSON(w, http.StatusOK, toUserResponse(user))
}

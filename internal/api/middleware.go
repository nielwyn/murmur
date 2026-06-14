package api

import (
	"context"
	"net/http"

	"murmur/internal/auth"
	"murmur/internal/database"
)

type contextKey string

const userContextKey contextKey = "user"

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetAuthCookie(r)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		userID, err := auth.ValidateJWT(token, s.cfg.JWTSecret)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		user, err := s.db.GetUserByID(r.Context(), userID)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

func userFromContext(r *http.Request) database.User {
	return r.Context().Value(userContextKey).(database.User)
}

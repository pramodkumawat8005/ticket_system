package main

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"
const usernameContextKey contextKey = "username"


func (app *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			writeError(w, http.StatusUnauthorized, "Authorization header must use Bearer scheme")
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if tokenString == "" {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := ParseToken(app.jwtSecret, tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID())
		ctx = context.WithValue(ctx, usernameContextKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func userIDFromContext(r *http.Request) string {
	id, _ := r.Context().Value(userIDContextKey).(string)
	return id
}

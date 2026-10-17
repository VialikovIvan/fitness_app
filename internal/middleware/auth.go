package middleware

import (
	"context"
	"net/http"
	"strings"
)

func (m *Manager) AuthOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("authToken")
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		token := cookie.Value

		userID, role, err := m.tokenValidator.ValidateAuth(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "role", role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *Manager) GuestOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("authToken")
		if err == nil {
			token := cookie.Value

			_, _, err := m.tokenValidator.ValidateAuth(r.Context(), token)
			if err == nil {
				http.Error(w, "Already authenticated", http.StatusForbidden)
				return 
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Manager) RequireBearer(next http.Handler, key string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		gotKey := strings.TrimPrefix(auth, prefix)
		if gotKey != key { 
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

        next.ServeHTTP(w, r)
    })
}

func GetUserID(r *http.Request) (int64, bool) {
	rawUserID := r.Context().Value("userID")
	if rawUserID == nil {
		return 0, false
	}
	userID, ok := rawUserID.(int64)
	return userID, ok
}

func GetRole(r *http.Request) (string, bool) {
	rawRole := r.Context().Value("role")
	if rawRole == nil {
		return "", false
	}
	role, ok := rawRole.(string)
	return role, ok
}

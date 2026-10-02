package middleware

import (
	"context"
	"net/http"
)

func (m *Manager) AuthOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("authToken")
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		token := cookie.Value

		userID, role, err := m.tokenValidator.Validate(r.Context(), token)
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

			_, _, err := m.tokenValidator.Validate(r.Context(), token)
			if err == nil {
				http.Error(w, "Already authenticated", http.StatusForbidden)
				return 
			}
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

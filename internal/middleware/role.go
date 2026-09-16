package middleware

import "net/http"

func (m *Manager) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRole(r)
		if !ok {
			m.logger.Error("failed to extract role")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetRole(r *http.Request) (string, bool) {
	rawRole := r.Context().Value("role")
	if rawRole == nil {
		return "", false
	}
	role, ok := rawRole.(string)
	return role, ok
}

package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func(m *Manager) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				m.logger.Error("recovered from panic", slog.Any("panic", err), slog.String("stack", string(debug.Stack())))

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
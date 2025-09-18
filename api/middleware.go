package api

import (
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type MiddlewareFunc func(http.Handler) http.Handler

func (s *Server) chainMiddleware(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	ms := []MiddlewareFunc{
		s.newAuthMiddleware(),
	}
	ms = append(ms, middlewares...)
	for _, mw := range ms {
		h = mw(h)
	}
	return h
}

func (s *Server) newAuthMiddleware() MiddlewareFunc {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			s.logger.Info("New Auth Token", zap.String("token", token))
			next.ServeHTTP(w, r)
		})
	}
}

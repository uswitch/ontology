package middleware

import (
	"fmt"
	"net/http"
)

type CORSConfig struct {
	AllowedOrigins []string
	MaxAge         int
}

type corsMiddleware struct {
	config CORSConfig
}

func NewCORSMiddleware(config CORSConfig) Middleware {
	return &corsMiddleware{
		config: config,
	}
}

func (m *corsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		responseOrigin := ""

		for _, allowedOrigin := range m.config.AllowedOrigins {
			if allowedOrigin == origin {
				responseOrigin = origin
				break
			}
		}

		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Origin", responseOrigin)
			w.Header().Add("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Add("Access-Control-Allow-Method", "GET, POST")
			w.Header().Add("Access-Control-Allow-Max-Age", fmt.Sprintf("%d", m.config.MaxAge))

			w.WriteHeader(204)
		} else {
			w.Header().Add("Access-Control-Allow-Origin", responseOrigin)

			next.ServeHTTP(w, r)
		}
	})
}

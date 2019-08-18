package audit

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type AuditEntry struct {
	User   string
	Method string
	Path   string
	Query  string
	Time   time.Time
}

func AuditMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(string)
		if !ok || user == "" {
			// we should always be after the auth middleware as we need to know about who's doing the action
			// so if we can't get the  user back something has gone a little pair shaped
			w.WriteHeader(500)
			return
		}

		byteString, err := json.Marshal(AuditEntry{
			User:   user,
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Time:   time.Now(),
		})

		if err != nil {
			log.Printf("Failed to marshal JSON: %v", err)
			w.WriteHeader(500)
			return
		}

		logger.Println(string(byteString))

		next.ServeHTTP(w, r)
	})
}

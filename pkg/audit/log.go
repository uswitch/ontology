package audit

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/uswitch/ontology/pkg/authnz"
	"github.com/uswitch/ontology/pkg/middleware"
)

const CallIDContextKey = "audit-call"

type AuditData map[string]interface{}

type AuditEntry struct {
	User   string
	CallID uuid.UUID
	Time   time.Time
	Data   AuditData
}

type Logger interface {
	middleware.Middleware

	Log(context.Context, AuditData)
}

type auditLog struct {
	logger *log.Logger
}

func NewAuditLogger(logger *log.Logger) Logger {
	return &auditLog{
		logger: logger,
	}
}

func (al *auditLog) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(authnz.UserContextKey).(string)
		if !ok || user == "" {
			// we should always be after the auth middleware as we need to know about who's doing the action
			// so if we can't get the  user back something has gone a little pair shaped
			w.WriteHeader(500)
			return
		}

		callID, err := uuid.NewRandom()
		if err != nil {
			log.Println("Failed to generate random audit call id")
			w.WriteHeader(500)
			return
		}

		newContext := context.WithValue(r.Context(), CallIDContextKey, callID)

		al.Log(newContext, AuditData{
			"origin": r.Header.Get("Origin"),
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		})

		next.ServeHTTP(w, r.WithContext(newContext))
	})
}

func (al *auditLog) Log(ctx context.Context, data AuditData) {
	user := ctx.Value(authnz.UserContextKey).(string)
	callID := ctx.Value(CallIDContextKey).(uuid.UUID)

	byteString, err := json.Marshal(AuditEntry{
		User:   user,
		CallID: callID,
		Time:   time.Now(),
		Data:   data,
	})

	if err != nil {
		log.Printf("Failed to marshal audit JSON: %v", err)
	}

	al.logger.Println(string(byteString))
}

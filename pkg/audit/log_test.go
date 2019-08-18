package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uswitch/ontology/pkg/authnz"
)

func auditTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}

func doAuditMiddleware(t *testing.T, expectedStatus int, expectedEntry AuditEntry, method, path, query string) {
	stringBuffer := &strings.Builder{}
	logger := log.New(stringBuffer, "", 0)

	req := httptest.NewRequest(
		method, fmt.Sprintf("%s?%s", path, query), nil,
	)
	reqWithUser := req.WithContext(context.WithValue(req.Context(), authnz.UserContextKey, expectedEntry.User))

	w := httptest.NewRecorder()

	auditLogger := NewAuditLog(logger)
	middleware := auditLogger.Middleware(auditTestHandler())
	middleware.ServeHTTP(w, reqWithUser)

	response := w.Result()

	if response.StatusCode != expectedStatus {
		t.Errorf("Should have got a %d, but got a %d", expectedStatus, response.StatusCode)
	}

	if response.StatusCode < 400 {
		loggerOutput := stringBuffer.String()

		var entry AuditEntry
		if err := json.Unmarshal([]byte(loggerOutput), &entry); err != nil {
			t.Fatalf("Couldn't decode entry from '%s': %v", loggerOutput, err)
		}

		if entry.User != expectedEntry.User {
			t.Errorf("should have been '%s', but was '%s'", expectedEntry.User, entry.User)
		}
		if now := time.Now(); now.Sub(entry.Time) > (1 * time.Second) {
			t.Errorf("should have been within a seconds of '%s', but was '%s'", now, entry.Time)
		}

		if len(entry.Data) != len(expectedEntry.Data) {
			t.Errorf("expected entry doesn 't have the same number of keys: %d != %d", len(entry.Data), len(expectedEntry.Data))
		}

		for k, v := range expectedEntry.Data {
			if entry.Data[k] != v {
				t.Errorf("Data['%s'] should have been '%v', but was '%v'", k, v, entry.Data[k])
			}
		}
	}

}

func TestAuditHappyPath(t *testing.T) {
	doAuditMiddleware(t, 200, AuditEntry{
		User: "wibble@bibble.com",
		Data: AuditData{
			"method": "GET",
			"path":   "/",
			"query":  "",
		},
	}, "GET", "/", "")
}

func TestMissingUser(t *testing.T) {
	doAuditMiddleware(t, 500, AuditEntry{
		User: "",
		Data: AuditData{
			"method": "GET",
			"path":   "/",
			"query":  "",
		},
	}, "GET", "/", "")
}

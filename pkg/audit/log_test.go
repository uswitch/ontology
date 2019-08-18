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
)

func auditTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}

func doAuditMiddleware(t *testing.T, expectedStatus int, expectedEntry AuditEntry) {
	stringBuffer := &strings.Builder{}
	logger := log.New(stringBuffer, "", 0)

	req := httptest.NewRequest(
		expectedEntry.Method,
		fmt.Sprintf("%s?%s", expectedEntry.Path, expectedEntry.Query),
		nil,
	)
	reqWithUser := req.WithContext(context.WithValue(req.Context(), "user", expectedEntry.User))

	w := httptest.NewRecorder()

	middleware := AuditMiddleware(logger, auditTestHandler())
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
		if entry.Method != expectedEntry.Method {
			t.Errorf("should have been '%s', but was '%s'", expectedEntry.Method, entry.Method)
		}
		if entry.Path != expectedEntry.Path {
			t.Errorf("should have been '%s', but was '%s'", expectedEntry.Path, entry.Path)
		}
		if entry.Query != expectedEntry.Query {
			t.Errorf("should have been '%s', but was '%s'", expectedEntry.Query, entry.Query)
		}
		if now := time.Now(); now.Sub(entry.Time) > (1 * time.Second) {
			t.Errorf("should have been within a seconds of '%s', but was '%s'", now, entry.Time)
		}
	}

}

func TestAuditHappyPath(t *testing.T) {
	doAuditMiddleware(t, 200, AuditEntry{
		User:   "wibble@bibble.com",
		Method: "GET",
		Path:   "/",
		Query:  "",
	})
}

func TestMissingUser(t *testing.T) {
	doAuditMiddleware(t, 500, AuditEntry{
		User:   "",
		Method: "GET",
		Path:   "/",
		Query:  "",
	})
}

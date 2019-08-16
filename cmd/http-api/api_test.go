package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/uswitch/ontology/pkg/audit"
	"github.com/uswitch/ontology/pkg/authnz"
	"github.com/uswitch/ontology/pkg/authnz/authnztest"
	"github.com/uswitch/ontology/pkg/middleware"
	"github.com/uswitch/ontology/pkg/store"
	"github.com/uswitch/ontology/pkg/store/inmem"
)

var (
	expectedUser   = "wibble@bibble.com"
	keys, token, _ = authnztest.SetupKeysAndToken(expectedUser, "https://bibble.com", "api", "sub")
	providerConfig = authnz.OIDCConfig{
		URL:       "https://bibble.com",
		Keys:      keys,
		ClientID:  "api",
		UserClaim: "sub",
	}
)

func thingWithType(thingID string, typeID string) *store.Thing {
	return &store.Thing{
		Metadata: store.Metadata{
			ID:   store.ID(thingID),
			Type: store.ID(typeID),
		},
	}
}
func entity(ID string) *store.Thing   { return thingWithType(ID, "/entity") }
func relation(ID string) *store.Thing { return thingWithType(ID, "/relation") }
func ntype(ID string) *store.Thing    { return thingWithType(ID, "/type") }

func doAPIRequest(srv *httptest.Server, token, method, path, body string) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", srv.URL, path), strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return srv.Client().Do(req)
}

func newAPIServer() (*httptest.Server, store.Store, error) {
	s := inmem.NewInMemoryStore()

	providers := []authnz.OIDCConfig{providerConfig}
	oidcAuth, err := authnz.NewOIDCAuthenticator(context.Background(), providers)
	if err != nil {
		return nil, nil, err
	}

	auditLogger := audit.NewAuditLogger(log.New(os.Stderr, "audit\t", 0))

	api, err := apiHandler(s, oidcAuth, auditLogger, middleware.PassThru)
	if err != nil {
		return nil, nil, err
	}

	return httptest.NewServer(api), s, nil
}

func TestAPIPost(t *testing.T) {
	srv, s, err := newAPIServer()
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	numBeforePOST, err := s.Len()
	if err != nil {
		t.Fatalf("Couldn't count numnber of things before post: %v", err)
	}

	thingJson, err := json.Marshal(entity("/wibble"))
	if err != nil {
		t.Fatalf("Couldn't marshal thing into JSON: %v", err)
	}

	if postResponse, err := doAPIRequest(srv, token, "POST", "/", string(thingJson)); err != nil {
		t.Fatalf("Failed to POST thing: %v", err)
	} else if expected := 200; postResponse.StatusCode != expected {
		t.Errorf("POST expected %d, but got %d", expected, postResponse.StatusCode)
	}

	numAfterPOST, err := s.Len()
	if err != nil {
		t.Fatalf("Couldn't count numnber of things after post: %v", err)
	}

	if numAfterPOST != (numBeforePOST + 1) {
		t.Errorf("Expected there to be one more thing than there was to start off with: %d != (%d + 1)", numAfterPOST, numBeforePOST)
	}
}

func TestAPIPostRawJSON(t *testing.T) {
	srv, s, err := newAPIServer()
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	thingJson := `
{
  "metadata": {
    "id": "/wibble",
    "type": "/entity"
  },
  "properties": {
    "wibble": "bibble"
  }
}
`

	if postResponse, err := doAPIRequest(srv, token, "POST", "/", thingJson); err != nil {
		t.Fatalf("Failed to POST thing: %v", err)
	} else if expected := 200; postResponse.StatusCode != expected {
		t.Errorf("POST expected %d, but got %d", expected, postResponse.StatusCode)
	}

	if entity, err := s.GetEntityByID(store.ID("/wibble")); err != nil {
		t.Errorf("Failed to get /wibble: %v", err)
	} else if entity.Metadata.Type != store.ID("/entity") {
		t.Errorf("/wibble has the wrong type: %v", entity)
	} else if entity.Properties["wibble"] != "bibble" {
		t.Errorf("/wibble had the wrong properties entry: %v", entity)
	}
}

func TestAPIPostWrongMethod(t *testing.T) {
	srv, _, err := newAPIServer()
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	thingJson, err := json.Marshal(entity("/wibble"))
	if err != nil {
		t.Fatalf("Couldn't marshal thing into JSON: %v", err)
	}

	allowedMethods := []string{http.MethodPost, http.MethodPut}

	for _, method := range allowedMethods {
		if postResponse, err := doAPIRequest(srv, token, method, "/", string(thingJson)); err != nil {
			t.Fatalf("Failed to %s thing: %v", method, err)
		} else if expected := 200; postResponse.StatusCode != expected {
			t.Errorf("%s expected %d, but got %d", method, expected, postResponse.StatusCode)
		}
	}

	disallowedMethods := []string{
		http.MethodGet, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace,
		// http.MethodHead, This won't send the body so will come through as a 400
	}

	for _, method := range disallowedMethods {
		if postResponse, err := doAPIRequest(srv, token, method, "/", string(thingJson)); err != nil {
			t.Fatalf("Failed to %s thing: %v", method, err)
		} else if expected := 405; postResponse.StatusCode != expected {
			t.Errorf("%s expected %d, but got %d", method, expected, postResponse.StatusCode)
		}
	}
}

func TestAPIPostNonsenseGetBadRequest(t *testing.T) {
	srv, _, err := newAPIServer()
	if err != nil {
		t.Fatalf("Failed to create API server: %v", err)
	}

	if postResponse, err := doAPIRequest(srv, token, "POST", "/", "nonsense"); err != nil {
		t.Fatalf("Failed to POST thing: %v", err)
	} else if expected := 400; postResponse.StatusCode != expected {
		t.Errorf("POST expected %d, but got %d", expected, postResponse.StatusCode)
	}
}

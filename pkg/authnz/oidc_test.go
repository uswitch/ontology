package authnz

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uswitch/ontology/pkg/authnz/authnztest"
)

var (
	expectedUser = "wibble@bibble.com"
	keys, token, _ = authnztest.SetupKeysAndToken(expectedUser, "https://bibble.com", "api", "sub")
	providerConfig = OIDCConfig{
		URL: "https://bibble.com",
		Keys: keys,
		ClientID: "api",
		UserClaim: "sub",
	}
	keys2, token2, _ = authnztest.SetupKeysAndToken(expectedUser, "https://thing.bibble.com", "thing", "sub")
	providerConfig2 = OIDCConfig{
		URL: "https://thing.bibble.com",
		Keys: keys2,
		ClientID: "thing",
		UserClaim: "sub",
	}
)

func TestOIDCHappyPath(t *testing.T) {
	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig}, fmt.Sprintf("Bearer %s", token))

	if response.StatusCode != 200 {
		t.Errorf("%d expected but got %d", 200, response.StatusCode)
	}

	if user != expectedUser {
		t.Errorf("'%s' expected, but got '%s'", expectedUser, user)
	}
}

func TestOIDCTwoProviders(t *testing.T) {

	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig2, providerConfig2}, fmt.Sprintf("Bearer %s", token2))

	if response.StatusCode != 200 {
		t.Errorf("%d expected but got %d", 200, response.StatusCode)
	}

	if user != expectedUser {
		t.Errorf("'%s' expected, but got '%s'", expectedUser, user)
	}
}

func TestOIDCNoAuthHeader(t *testing.T) {
	response, _ := doOIDCMiddleware(t, []OIDCConfig{}, "")

	if response.StatusCode != 401 {
		t.Errorf("%d expected but got %d", 401, response.StatusCode)
	}
}

func TestOIDCAuthHeaderMalformed(t *testing.T) {
	headers := []string{
		"Basic af54hhrd",
		"   ",
		"bearer sgerg",
		"token   ",
	}

	for _, header := range headers {
		response, _ := doOIDCMiddleware(t, []OIDCConfig{}, header)

		if response.StatusCode != 401 {
			t.Errorf("%d expected but got %d", 401, response.StatusCode)
		}
	}
}

func TestOIDCNoVerifiedProvider(t *testing.T) {
	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig}, fmt.Sprintf("Bearer %s", token2))

	if response.StatusCode != 401 {
		t.Errorf("%d expected but got %d", 401, response.StatusCode)
	}

	if user == expectedUser {
		t.Error("user shouldn't be correct")
	}
}

func TestOIDCNoMatchingUserClaim(t *testing.T) {
	expectedUser := "wibble@bibble.com"
	keys, token, err := authnztest.SetupKeysAndToken(expectedUser, "https://bibble.com", "api", "user")
	if err != nil {
		t.Fatalf("Couldn't create keys and token: %v", err)
	}
	providerConfig = OIDCConfig{
		URL: "https://bibble.com",
		Keys: keys,
		ClientID: "api",
		UserClaim: "user",
	}


	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig}, fmt.Sprintf("Bearer %s", token))

	if response.StatusCode != 500 {
		t.Errorf("%d expected but got %d", 500, response.StatusCode)
	}

	if user == expectedUser {
		t.Error("user shouldn't be correct")
	}
}

func oidcTestHandler(out *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*out = r.Context().Value(UserContextKey).(string)
		w.WriteHeader(200)
	})
}

func doOIDCMiddleware(t *testing.T, config []OIDCConfig, authorizationHeader string) (*http.Response, string) {
	authenticator, err := NewOIDCAuthenticator(context.Background(), config)
	if err != nil {
		t.Fatalf("Couldn't create the authenticator: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	if authorizationHeader != "" {
		req.Header.Add("Authorization", authorizationHeader)
	}

	w := httptest.NewRecorder()

	var user string
	authenticator.Middleware(oidcTestHandler(&user)).ServeHTTP(w, req)

	// expect use to be populated with the correct username
	response := w.Result()

	return response, user
}

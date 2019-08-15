package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestOIDCHappyPath(t *testing.T) {
	// given a provider config and token, for a known username
	expectedUser := "wibble@bibble.com"
	providerConfig, token, err := setupProviderAndToken(expectedUser, "https://bibble.com", "api", "sub")
	if err != nil {
		t.Fatalf("Couldn't create provider and token: %v", err)
	}

	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig}, fmt.Sprintf("Bearer %s", token))

	if response.StatusCode != 200 {
		t.Errorf("%d expected but got %d", 200, response.StatusCode)
	}

	if user != expectedUser {
		t.Errorf("'%s' expected, but got '%s'", expectedUser, user)
	}
}

func TestOIDCTwoProviders(t *testing.T) {
	expectedUser := "wibble@bibble.com"
	providerConfig1, _, err := setupProviderAndToken(expectedUser, "https://bibble.com", "api", "sub")
	if err != nil {
		t.Fatalf("Couldn't create provider1 and token: %v", err)
	}
	providerConfig2, token, err := setupProviderAndToken(expectedUser, "https://thing.bibble.com", "thing", "sub")
	if err != nil {
		t.Fatalf("Couldn't create provider2 and token: %v", err)
	}

	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig1, providerConfig2}, fmt.Sprintf("Bearer %s", token))

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
	expectedUser := "wibble@bibble.com"
	providerConfig, _, err := setupProviderAndToken(expectedUser, "https://bibble.com", "api", "sub")
	if err != nil {
		t.Fatalf("Couldn't create provider1 and token: %v", err)
	}
	_, token, err := setupProviderAndToken(expectedUser, "https://thing.bibble.com", "thing", "sub")
	if err != nil {
		t.Fatalf("Couldn't create provider2 and token: %v", err)
	}

	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig}, fmt.Sprintf("Bearer %s", token))

	if response.StatusCode != 401 {
		t.Errorf("%d expected but got %d", 401, response.StatusCode)
	}

	if user == expectedUser {
		t.Error("user shouldn't be correct")
	}
}

func TestOIDCNoMatchingUserClaim(t *testing.T) {
	expectedUser := "wibble@bibble.com"
	providerConfig, token, err := setupProviderAndToken(expectedUser, "https://bibble.com", "api", "user")
	if err != nil {
		t.Fatalf("Couldn't create provider1 and token: %v", err)
	}

	response, user := doOIDCMiddleware(t, []OIDCConfig{providerConfig}, fmt.Sprintf("Bearer %s", token))

	if response.StatusCode != 500 {
		t.Errorf("%d expected but got %d", 500, response.StatusCode)
	}

	if user == expectedUser {
		t.Error("user shouldn't be correct")
	}
}

func setupProviderAndToken(user, iss, aud, claim string) (OIDCConfig, string, error) {
	kid := "wibble"
	alg := "RS256"
	use := "sig"

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return OIDCConfig{}, "", err
	}

	priv := jose.JSONWebKey{Key: key, KeyID: kid, Algorithm: alg, Use: use}
	pub := jose.JSONWebKey{Key: key.Public(), KeyID: kid, Algorithm: alg, Use: use}

	config := OIDCConfig{
		URL:       iss,
		Keys:      []jose.JSONWebKey{priv, pub},
		ClientID:  aud,
		UserClaim: claim,
	}

	sig, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key:       priv,
		},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return OIDCConfig{}, "", err
	}

	cl := jwt.Claims{
		Subject:   user,
		Issuer:    iss,
		Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		NotBefore: jwt.NewNumericDate(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)),
		Audience:  jwt.Audience{aud},
	}
	raw, err := jwt.Signed(sig).Claims(cl).CompactSerialize()
	if err != nil {
		return OIDCConfig{}, "", err
	}

	return config, raw, nil
}

func oidcTestHandler(out *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*out = r.Context().Value("user").(string)
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

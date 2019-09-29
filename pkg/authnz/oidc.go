package authnz

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"gopkg.in/square/go-jose.v2"
)

type OIDCConfig struct {
	URL  string
	Keys []jose.JSONWebKey

	ClientID  string
	UserClaim string
}

type provider struct {
	Config   *OIDCConfig
	Verifier *oidc.IDTokenVerifier
}

type OIDC struct {
	providers []*provider
}

type LocalKeySet []jose.JSONWebKey

func (keys LocalKeySet) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	jws, err := jose.ParseSigned(jwt)
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt: %v", err)
	}

	keyID := ""
	for _, sig := range jws.Signatures {
		keyID = sig.Header.KeyID
		break
	}

	for _, key := range keys {
		if keyID == "" || key.KeyID == keyID {
			if payload, err := jws.Verify(&key); err == nil {
				return payload, nil
			}
		}
	}

	return nil, errors.New("failed to verify id token signature")
}

func NewOIDCAuthenticator(ctx context.Context, providerConfigs []OIDCConfig) (Authenticator, error) {
	oidcConfig := OIDC{
		providers: make([]*provider, len(providerConfigs)),
	}

	for idx, providerConfig := range providerConfigs {
		var verifier *oidc.IDTokenVerifier

		verifierConfig := &oidc.Config{ClientID: providerConfig.ClientID}

		if len(providerConfig.Keys) > 0 {
			keySet := LocalKeySet(providerConfig.Keys)
			verifier = oidc.NewVerifier(providerConfig.URL, keySet, verifierConfig)
		} else {
			p, err := oidc.NewProvider(ctx, providerConfig.URL)
			if err != nil {
				return nil, err
			}

			verifier = p.Verifier(verifierConfig)
		}

		oidcConfig.providers[idx] = &provider{
			Config:   &providerConfig,
			Verifier: verifier,
		}
	}

	return &oidcConfig, nil
}

func (o *OIDC) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header, ok := r.Header["Authorization"]

		if !ok {
			log.Println("No Authorization header found")
			w.WriteHeader(401)
			return
		}

		headerParts := strings.Split(header[0], " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			log.Printf("Expected Authorization header to contain Bearer token, but was: %s", headerParts[0])
			w.WriteHeader(401)
		}

		rawIDToken := headerParts[1]

		var idToken *oidc.IDToken
		var err error
		var verifiedProvider *provider

		for _, provider := range o.providers {
			idToken, err = provider.Verifier.Verify(r.Context(), rawIDToken)
			if err == nil {
				verifiedProvider = provider
				break
			}
		}

		if verifiedProvider == nil {
			log.Println("Failed to find a provider that could validate the token")
			w.WriteHeader(401)
			return
		}

		claims := map[string]interface{}{}

		err = idToken.Claims(&claims)
		if err != nil {
			log.Println("Failed to dump the claims into a map")
			w.WriteHeader(500)
			return
		}

		user, ok := claims[verifiedProvider.Config.UserClaim]
		if !ok {
			log.Printf("Couldn't extract user from token, claim '%s' doesn't exist", verifiedProvider.Config.UserClaim)
			w.WriteHeader(500)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey, user)))
	})
}

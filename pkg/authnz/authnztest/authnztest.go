package authnztest

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)


func SetupKeysAndToken(user, iss, aud, claim string) ([]jose.JSONWebKey, string, error) {
	kid := "wibble"
	alg := "RS256"
	use := "sig"

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	priv := jose.JSONWebKey{Key: key, KeyID: kid, Algorithm: alg, Use: use}
	pub := jose.JSONWebKey{Key: key.Public(), KeyID: kid, Algorithm: alg, Use: use}

	sig, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key:       priv,
		},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	}

	return []jose.JSONWebKey{priv, pub}, raw, nil
}

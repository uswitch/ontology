package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"gopkg.in/square/go-jose.v2"
)

type OIDCConfig struct {
	URL  string
	Keys []jose.JSONWebKey

	ClientID  string
	UserClaim string
}

type Config struct {
	ApiAddr string
	OpsAddr string

	Providers []OIDCConfig
}

var config = Config{
	ApiAddr: "127.0.0.1:8080",
	OpsAddr: "127.0.0.1:8081",
}

func (c Config) validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("You need to have at least one OIDC provider defined")
	}

	for _, provider := range c.Providers {
		if _, err := url.Parse(provider.URL); err != nil {
			return fmt.Errorf("%v has an invalid URL: %v", provider, err)
		}

		if provider.ClientID == "" {
			return fmt.Errorf("%v needs a client id", provider)
		}

		if provider.UserClaim == "" {
			provider.UserClaim = "sub"
		}
	}

	return nil
}

func ConfigFromPath(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

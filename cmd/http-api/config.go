package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/uswitch/ontology/pkg/authnz"
	"github.com/uswitch/ontology/pkg/middleware"
)

type ServerConfig struct {
	Addr string

	WriteTimeoutSecs uint
	ReadTimeoutSecs  uint
	IdleTimeoutSecs  uint
}

type WSConfig struct {
	ReadBufferSize  int
	WriteBufferSize int

	AllowedOrigins []string
}

type ApiConfig struct {
	Server ServerConfig
	CORS   middleware.CORSConfig
	WS     WSConfig
}

type OpsConfig struct {
	Server ServerConfig
}

type Config struct {
	Api ApiConfig
	Ops OpsConfig

	GracefulTimeoutSecs uint

	Providers []authnz.OIDCConfig
}

var config = Config{
	GracefulTimeoutSecs: 15,
	Api: ApiConfig{
		Server: ServerConfig{
			Addr: "127.0.0.1:8080",

			WriteTimeoutSecs: 15,
			ReadTimeoutSecs:  15,
			IdleTimeoutSecs:  60,
		},
		CORS: middleware.CORSConfig{
			AllowedOrigins: []string{},
			MaxAge:         86400, // 24 hours
		},
		WS: WSConfig{
			AllowedOrigins:  []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	},
	Ops: OpsConfig{
		Server: ServerConfig{
			Addr: "127.0.0.1:8081",

			WriteTimeoutSecs: 15,
			ReadTimeoutSecs:  15,
			IdleTimeoutSecs:  60,
		},
	},
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

	for _, origin := range c.Api.CORS.AllowedOrigins {
		if _, err := url.Parse(origin); err != nil {
			return fmt.Errorf("%v has an invalid URL '%s': %v", c.Api.CORS.AllowedOrigins, origin, err)
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

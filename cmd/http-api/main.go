package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func apiHandler(config *Config) (http.Handler, error) {
	oidcAuth, err := NewOIDCAuthenticator(context.Background(), config.Providers)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load OIDC providers: %v", err)
	}

	auditLogger := log.New(os.Stderr, "audit\t", 0)

	apiMux := http.NewServeMux()

	return oidcAuth.Middleware(AuditMiddleware(auditLogger, apiMux)), nil
}

func opsHandler(config *Config) (http.Handler, error) {
	opsMux := http.NewServeMux()

	opsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "OK")
	})

	return opsMux, nil
}

func startServerInGroup(wg *sync.WaitGroup, name string, handler http.Handler, config ServerConfig) *http.Server {
	server := &http.Server{
		Addr:    config.Addr,
		Handler: handler,

		WriteTimeout: time.Second * time.Duration(config.WriteTimeoutSecs),
		ReadTimeout:  time.Second * time.Duration(config.ReadTimeoutSecs),
		IdleTimeout:  time.Second * time.Duration(config.IdleTimeoutSecs),
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Printf(
			"%s server listening on %v. timeouts: [w %v, r %v, i %v]",
			name,
			server.Addr,
			server.WriteTimeout,
			server.ReadTimeout,
			server.IdleTimeout,
		)
		log.Println(server.ListenAndServe())
	}()

	return server
}

func main() {
	var serverWaitGroup sync.WaitGroup

	if len(os.Args) != 2 {
		log.Fatal("http-api [config path]")
	}

	configPath := os.Args[1]
	config, err := ConfigFromPath(configPath)
	if err != nil {
		log.Fatalf("Couldn't load config file from '%s': %v", configPath, err)
	}

	if api, err := apiHandler(config); err != nil {
		log.Fatal(err)
	} else {
		startServerInGroup(&serverWaitGroup, "api", api, config.Api)
	}

	if ops, err := opsHandler(config); err != nil {
		log.Fatal(err)
	} else {
		startServerInGroup(&serverWaitGroup, "ops", ops, config.Ops)
	}

	serverWaitGroup.Wait()

}

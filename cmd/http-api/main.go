package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
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
		apiServer := &http.Server{
			Addr:    config.ApiAddr,
			Handler: api,
		}

		serverWaitGroup.Add(1)

		go func() {
			defer serverWaitGroup.Done()

			log.Printf("API server listening on: %v", apiServer.Addr)
			log.Println(apiServer.ListenAndServe())
		}()
	}

	if ops, err := opsHandler(config); err != nil {
		log.Fatal(err)
	} else {
		opsServer := &http.Server{
			Addr:    config.OpsAddr,
			Handler: ops,
		}

		serverWaitGroup.Add(1)

		go func() {
			defer serverWaitGroup.Done()

			log.Printf("Ops server listening on: %v", opsServer.Addr)
			log.Println(opsServer.ListenAndServe())
		}()
	}

	serverWaitGroup.Wait()

}

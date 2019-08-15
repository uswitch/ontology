package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

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

	oidcAuth, err := NewOIDCAuthenticator(context.Background(), config.Providers)
	if err != nil {
		log.Fatalf("Couldn't load OIDC providers: %v", err)
	}

	auditLogger := log.New(os.Stderr, "audit\t", 0)

	apiMux := http.NewServeMux()

	apiServer := &http.Server{
		Addr:    config.ApiAddr,
		Handler: oidcAuth.Middleware(AuditMiddleware(auditLogger, apiMux)),
	}

	serverWaitGroup.Add(1)

	go func() {
		defer serverWaitGroup.Done()

		log.Printf("API server listening on: %v", apiServer.Addr)
		log.Println(apiServer.ListenAndServe())
	}()

	opsMux := http.NewServeMux()

	opsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "OK")
	})

	opsServer := &http.Server{
		Addr:    config.OpsAddr,
		Handler: opsMux,
	}

	serverWaitGroup.Add(1)

	go func() {
		defer serverWaitGroup.Done()

		log.Printf("Ops server listening on: %v", opsServer.Addr)
		log.Println(opsServer.ListenAndServe())
	}()

	serverWaitGroup.Wait()

}

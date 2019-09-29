package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	"github.com/uswitch/ontology/pkg/audit"
	"github.com/uswitch/ontology/pkg/authnz"
	"github.com/uswitch/ontology/pkg/middleware"
	"github.com/uswitch/ontology/pkg/store/inmem"
)

func opsHandler() (http.Handler, error) {
	opsMux := http.NewServeMux()

	opsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "OK")
	})

	return opsMux, nil
}

func startServer(name string, handler http.Handler, config ServerConfig) *http.Server {
	server := &http.Server{
		Addr:    config.Addr,
		Handler: handler,

		WriteTimeout: time.Second * time.Duration(config.WriteTimeoutSecs),
		ReadTimeout:  time.Second * time.Duration(config.ReadTimeoutSecs),
		IdleTimeout:  time.Second * time.Duration(config.IdleTimeoutSecs),
	}

	go func() {
		log.Printf(
			"%s server listening on %v. timeouts: [w %v, r %v, i %v]",
			name,
			server.Addr,
			server.WriteTimeout,
			server.ReadTimeout,
			server.IdleTimeout,
		)

		err := server.ListenAndServe()

		switch err {
		case http.ErrServerClosed:
			log.Printf("%s server has shutdown", name)
		default:
			log.Fatalf("%s server failed: %v", name, err)
		}
	}()

	return server
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("http-api [config path]")
	}

	configPath := os.Args[1]
	config, err := ConfigFromPath(configPath)
	if err != nil {
		log.Fatalf("Couldn't load config file from '%s': %v", configPath, err)
	}

	var apiServer, opsServer *http.Server

	oidcAuth, err := authnz.NewOIDCAuthenticator(context.Background(), config.Providers)
	if err != nil {
		log.Fatalf("Couldn't load OIDC providers: %v", err)
	}

	auditLogger := audit.NewAuditLogger(log.New(os.Stderr, "audit\t", 0))
	cors := middleware.NewCORSMiddleware(config.Api.CORS)

	store := inmem.NewInMemoryStore()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.Api.WS.ReadBufferSize,
		WriteBufferSize: config.Api.WS.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			for _, allowedOrigin := range config.Api.WS.AllowedOrigins {
				if allowedOrigin == origin {
					return true
				}
			}

			return false
		},
		Subprotocols: []string{"graphql-ws"},
	}

	if api, err := apiHandler(store, upgrader, oidcAuth, auditLogger, cors); err != nil {
		log.Fatal(err)
	} else {
		apiServer = startServer("api", api, config.Api.Server)
	}

	if ops, err := opsHandler(); err != nil {
		log.Fatal(err)
	} else {
		opsServer = startServer("ops", ops, config.Ops.Server)
	}

	gracefulTimeout := time.Second * time.Duration(config.GracefulTimeoutSecs)
	gracefulShutdownSignals := make(chan os.Signal, 1)

	signal.Notify(gracefulShutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	<-gracefulShutdownSignals

	log.Printf("graceful shutdown triggered, waiting for %v for servers to shutdown", gracefulTimeout)

	wg := sync.WaitGroup{}
	shutdownDone := make(chan struct{})

	wg.Add(2)
	go func() { wg.Wait(); close(shutdownDone) }()

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	go func() { apiServer.Shutdown(ctx); wg.Done() }()
	go func() { opsServer.Shutdown(ctx); wg.Done() }()

	select {
	case <-ctx.Done():
		log.Println("timed out waiting for servers to shutdown")
	case <-shutdownDone:
		log.Println("both servers shutdown successfully")
	}

	os.Exit(0)
}

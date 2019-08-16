package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/uswitch/ontology/pkg/audit"
	"github.com/uswitch/ontology/pkg/authnz"
	"github.com/uswitch/ontology/pkg/middleware"
	"github.com/uswitch/ontology/pkg/store"
)


func apiHandler(s store.Store, authn authnz.Authenticator, auditLogger audit.Logger, cors middleware.Middleware) (http.Handler, error) {
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if ! (r.Method == "POST" || r.Method == "PUT") {
			w.WriteHeader(405)
			return
		}

		decoder := json.NewDecoder(r.Body)

		for ;; {
			var thing store.Thing

			if err := decoder.Decode(&thing); err == io.EOF {
				break
			} else if err != nil {
				log.Printf("Couldn't unmarshal a thing from request body: %v", err)
				w.WriteHeader(400)
				return
			}

			if err := s.Add(&thing); err != nil {
				log.Printf("coudln't add thing to store: %v", err)
				w.WriteHeader(500)
				return
			}
		}
	})

	return middleware.Wrap(
		[]middleware.Middleware{
			cors,
			authn,
			auditLogger,
		},
		apiMux,
	), nil
}

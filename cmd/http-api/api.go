package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/graphql-go/handler"

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

		thingsAdded := []store.ID{}
		defer func() {
			auditLogger.Log(r.Context(), audit.AuditData{"thingsAdded": thingsAdded})
		}()

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

			thingsAdded = append(thingsAdded, thing.Metadata.ID)
		}
	})

	schema, err := NewGraphQLSchema(s)
	if err != nil {
		return nil, err
	}

	h := handler.New(&handler.Config{
		Schema: schema,
		Pretty: true,
		GraphiQL: false,
		Playground: true,
	})

	apiMux.Handle("/graphql", h)

	return middleware.Wrap(
		[]middleware.Middleware{
			cors,
			authn,
			auditLogger,
		},
		apiMux,
	), nil
}

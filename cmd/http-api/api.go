package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	graphqlgo "github.com/graphql-go/graphql"

	"github.com/uswitch/ontology/pkg/audit"
	"github.com/uswitch/ontology/pkg/authnz"
	"github.com/uswitch/ontology/pkg/graphql"
	"github.com/uswitch/ontology/pkg/middleware"
	"github.com/uswitch/ontology/pkg/store"
)

type RequestOptions struct {
    Query         string                 `json:"query" url:"query" schema:"query"`
    Variables     map[string]interface{} `json:"variables" url:"variables" schema:"variables"`
    OperationName string                 `json:"operationName" url:"operationName" schema:"operationName"`
}

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

	schema, err := graphql.NewSchema(s)
	if err != nil {
		return nil, err
	}

	apiMux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		var opts RequestOptions
		err = json.Unmarshal(bodyBytes, &opts)
		if err != nil {
			// try an array, it's what apollo sends
			var manyOpts []RequestOptions
			err = json.Unmarshal(bodyBytes, &manyOpts)
			if err != nil || len(manyOpts) == 0 {
				w.WriteHeader(400)
				return
			}

			opts = manyOpts[0]
		}

		ctx := r.Context()

		auditLogger.Log(ctx, audit.AuditData{
			"query": opts.Query,
			"variables": opts.Variables,
			"operation_name": opts.OperationName,
		})

		result := graphqlgo.Do(graphqlgo.Params{
			Schema:         *schema,
			RequestString:  opts.Query,
			VariableValues: opts.Variables,
			OperationName:  opts.OperationName,
			Context:        ctx,
		})

		json.NewEncoder(w).Encode(result)
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

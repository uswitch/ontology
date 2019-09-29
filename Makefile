
.PHONY: all test

all: bin/http-api bin/query

test:
	go test github.com/uswitch/ontology/pkg/audit
	go test github.com/uswitch/ontology/pkg/authnz
	go test github.com/uswitch/ontology/pkg/graphql
	go test github.com/uswitch/ontology/pkg/graphql/ws
	go test github.com/uswitch/ontology/pkg/store
	go test github.com/uswitch/ontology/pkg/store/inmem
	go test github.com/uswitch/ontology/cmd/http-api

bin/http-api: $(shell find cmd/http-api -iname '*.go') $(shell find pkg/ -iname '*.go')
	go build -o bin/http-api ./cmd/http-api

bin/query: $(shell find cmd/query -iname '*.go') $(shell find pkg/ -iname '*.go')
	go build -o bin/query ./cmd/query


.PHONY: all test

all: bin/http-api

test:
	go test github.com/uswitch/ontology/pkg/store

bin/http-api: $(shell find cmd/http-api -iname '*.go') $(shell find pkg/ -iname '*.go')
	go build -o bin/http-api ./cmd/http-api

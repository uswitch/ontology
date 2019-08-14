
.PHONY: all test

all: bin/http-api

test:
	go test github.com/uswitch/ontology/pkg/store
	go test github.com/uswitch/ontology/cmd/http-api

bin/http-api: $(shell find cmd/http-api -iname '*.go') $(shell find pkg/ -iname '*.go')
	go build -o bin/http-api ./cmd/http-api

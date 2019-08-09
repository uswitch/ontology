
.PHONY: all

all: bin/http-api

bin/http-api: $(find cmd/http-api -iname '*.go') $(find pkg/ -iname '*.go')
	go build -o bin/http-api ./cmd/http-api

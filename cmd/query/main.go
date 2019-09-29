package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	graphqlws "github.com/uswitch/ontology/pkg/graphql/ws"
)

var (
	url    = flag.String("url", "ws://localhost:8080/graphqlws", "url to connect to")
	origin = flag.String("origin", "cli://ontology-client", "origin to send to the server")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, *url, http.Header{
		"Origin": []string{*origin},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to dial %s: %v", *url, err)
		os.Exit(1)
	}

	client := graphqlws.NewClientChannel(ctx, conn)

	err = client.Connect(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to graphql server: %v", err)
		os.Exit(1)
	}

	results, err := client.Operation(ctx, graphqlws.OperationParams{
		Query: flag.Arg(0),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query graphql server: %v", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)

	for done := false; !done; {
		select {
		case result := <-results:
			if result == nil {
				done = true
				break
			}

			enc.Encode(result)
		}
	}

}

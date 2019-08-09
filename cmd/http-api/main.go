package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {

	var serverWaitGroup sync.WaitGroup

	apiMux := http.NewServeMux()

	apiServer := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: apiMux,
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
		Addr:    "127.0.0.1:8081",
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

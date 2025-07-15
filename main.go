package main

import (
	"log"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	handler := http.FileServer(http.Dir("."))
	serveMux.Handle("/", handler)
	log.Fatal(server.ListenAndServe())
}

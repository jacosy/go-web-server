package main

import (
	"log"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	// Serve static files from the root directory
	serveMux.Handle("/", http.FileServer(http.Dir(".")))

	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

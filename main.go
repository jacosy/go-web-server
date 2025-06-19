package main

import (
	"log"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	// Serve static files from the root directory
	prefixHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := &apiConfig{}
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(prefixHandler))

	serveMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	serveMux.HandleFunc("GET /api/metrics", apiCfg.MetricsHandler)
	serveMux.HandleFunc("POST /api/reset", apiCfg.ResetMetricsHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

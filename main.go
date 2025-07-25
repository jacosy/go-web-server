package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Import PostgreSQL driver

	"github.com/jacosy/go-web-server/handler"
	"github.com/jacosy/go-web-server/internal/database"
)

func main() {
	// Load environment variables from .env file
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	dbQueries := database.New(db)
	secretKey := os.Getenv("SECRET_KEY")
	apiCfg := &apiConfig{
		db:        dbQueries,
		env:       os.Getenv("PLATFORM"),
		secretKey: secretKey,
	}

	serveMux := http.NewServeMux()
	// Serve static files from the root directory
	prefixHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(prefixHandler))

	serveMux.HandleFunc("GET /admin/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.ResetMetricsHandler)

	serveMux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	serveMux.HandleFunc("POST /api/login", apiCfg.LoginUser)

	chirpHandler := handler.NewChirpHandler(dbQueries, secretKey)
	serveMux.HandleFunc("POST /api/chirps", chirpHandler.CreateChirp)
	serveMux.HandleFunc("GET /api/chirps", chirpHandler.GetChirps)
	serveMux.HandleFunc("GET /api/chirps/{id}", chirpHandler.GetChirpByID)

	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

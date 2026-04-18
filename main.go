package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sorinfratila/chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// can also be done like this
	// s := http.Server{}
	// s.Addr = ":8080"
	// s.Handler = mux

	fileServerRoot := http.FileServer(http.Dir(filepathRoot))
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServerRoot))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	assetsServer := http.FileServer(http.Dir("./assets"))
	mux.Handle("assets/", assetsServer)

	// mux.HandleFunc("POST /api/users", handleUserCreation)
	//
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

// func (apiCfg *apiConfig) handleUserCreation(w http.ResponseWriter, r *http.Request) {
//
// 	user, err := apiCfg.db.CreateUser(r.Context(), params.Email)
// }

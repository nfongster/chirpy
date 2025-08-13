package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	f := func(wrt http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(wrt, req)
	}
	return http.HandlerFunc(f)
}

func getValidateChirpResponse(chirpLength int) (int, any) {
	if chirpLength > 140 {
		return 400, struct {
			Error string `json:"error"`
		}{
			Error: "Chirp is too long",
		}
	} else {
		return 200, struct {
			Valid bool `json:"valid"`
		}{
			Valid: true,
		}
	}
}

func main() {
	fmt.Println("Starting chirpy server...")

	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", func(wrt http.ResponseWriter, _ *http.Request) {
		wrt.Header().Set("Content-Type", "text/plain; charset=utf-8")
		wrt.WriteHeader(200)
		wrt.Write([]byte("OK\n"))
	})

	mux.HandleFunc("GET /admin/metrics", func(wrt http.ResponseWriter, _ *http.Request) {
		wrt.Header().Set("Content-Type", "text/html; charset=utf-8")
		wrt.WriteHeader(200)
		hits := apiCfg.fileserverHits.Load()
		html := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", hits)
		wrt.Write([]byte(html))
	})

	mux.HandleFunc("POST /admin/reset", func(wrt http.ResponseWriter, _ *http.Request) {
		wrt.Header().Set("Content-Type", "text/plain; charset=utf-8")
		wrt.WriteHeader(200)
		apiCfg.fileserverHits.Swap(0)
	})

	mux.HandleFunc("POST /api/validate_chirp", func(wrt http.ResponseWriter, req *http.Request) {
		// Handle request
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		if err := decoder.Decode(&params); err != nil {
			fmt.Printf("Error decoding parameters: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		// Write response
		code, jsonMessage := getValidateChirpResponse(len(params.Body))
		dat, err := json.Marshal(jsonMessage)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		wrt.Header().Set("Content-Type", "application/json")
		wrt.WriteHeader(code)
		wrt.Write(dat)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server failure.  Error: %v", err)
		os.Exit(1)
	}
}

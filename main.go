package main

import (
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

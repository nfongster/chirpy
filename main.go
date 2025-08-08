package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting chirpy server...")

	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	mux.HandleFunc("/healthz", func(wrt http.ResponseWriter, _ *http.Request) {
		wrt.Header().Set("Content-Type", "text/plain; charset=utf-8")
		wrt.WriteHeader(200)
		wrt.Write([]byte("OK"))
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

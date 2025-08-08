package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting chirpy server...")

	server := http.Server{
		Addr:    ":8080",
		Handler: http.NewServeMux(),
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server failure.  Error: %v", err)
		os.Exit(1)
	}
}

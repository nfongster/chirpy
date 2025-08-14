package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nfongster/chirpy/internal/auth"
	"github.com/nfongster/chirpy/internal/database"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	f := func(wrt http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(wrt, req)
	}
	return http.HandlerFunc(f)
}

func validateChirp(chirp string) bool {
	return len(chirp) <= 140
}

func removeProfanities(chirp string) string {
	profanities := []string{
		"kerfuffle", "sharbert", "fornax",
	}

	words := strings.Split(chirp, " ")
	for i, word := range strings.Split(chirp, " ") {
		if slices.Contains(profanities, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func main() {
	fmt.Println("Starting chirpy server...")

	if err := godotenv.Load(); err != nil {
		fmt.Printf("error loading db string: %v\n", err)
		os.Exit(1)
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("error opening db: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		db: dbQueries,
	}

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

	mux.HandleFunc("POST /admin/reset", func(wrt http.ResponseWriter, req *http.Request) {
		wrt.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if platform != "dev" {
			wrt.WriteHeader(403)
			return
		}

		apiCfg.fileserverHits.Swap(0)
		if err := apiCfg.db.DeleteAllUsers(req.Context()); err != nil {
			fmt.Printf("Error deleting all users from db: %s\n", err)
			wrt.WriteHeader(500)
			return
		}
		wrt.WriteHeader(200)
	})

	mux.HandleFunc("POST /api/users", func(wrt http.ResponseWriter, req *http.Request) {
		// Handle request
		decoder := json.NewDecoder(req.Body)
		params := userParameters{}
		if err := decoder.Decode(&params); err != nil {
			fmt.Printf("Error decoding parameters: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		// Write response
		if params.Password == "" {
			wrt.WriteHeader(400)
			wrt.Write([]byte("No password was supplied!"))
			return
		}
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			fmt.Printf("Error hashing password: %v", err)
			wrt.WriteHeader(500)
			return
		}
		user, err := apiCfg.db.CreateUser(req.Context(), database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			fmt.Printf("Error querying user for email %s: %s\n", params.Email, err)
			wrt.WriteHeader(500)
			return
		}

		// Convert DB query struct to JSON struct
		dat, err := json.Marshal(User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		})
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		wrt.Header().Set("Content-Type", "application/json")
		wrt.WriteHeader(201)
		wrt.Write(dat)
	})

	mux.HandleFunc("POST /api/login", func(wrt http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := userParameters{}
		if err := decoder.Decode(&params); err != nil {
			fmt.Printf("Error decoding parameters: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		user, err := apiCfg.db.GetUserByEmail(req.Context(), params.Email)
		if err != nil {
			wrt.WriteHeader(401)
			fmt.Fprint(wrt, "No user with the email %s was found", params.Email)
			return
		}
		// Check to see if requested password matches stored hash
		if err := auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
			wrt.WriteHeader(401)
			wrt.Write([]byte("incorrect email or password"))
			return
		}

		dat, err := json.Marshal(User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		})
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s\n", err)
			wrt.WriteHeader(500)
			return
		}
		wrt.WriteHeader(200)
		wrt.Write(dat)
	})

	mux.HandleFunc("POST /api/chirps", func(wrt http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := chirpParameters{}
		if err := decoder.Decode(&params); err != nil {
			fmt.Printf("Error decoding parameters: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		wrt.Header().Set("Content-Type", "application/json")
		if !validateChirp(params.Body) {
			dat, err := json.Marshal(chirpError{
				Error: "Chirp is too long",
			})
			if err != nil {
				fmt.Printf("Error marshalling JSON: %v\n", err)
				wrt.WriteHeader(500)
			} else {
				wrt.WriteHeader(400)
				wrt.Write(dat)
			}
			return
		}

		if _, err := apiCfg.db.GetUser(req.Context(), params.UserId); err != nil {
			fmt.Printf("Error retrieving user for id %v: %v", params.UserId, err)
			wrt.WriteHeader(400)
			wrt.Write([]byte(fmt.Sprintf("No user with id %v exists in the DB!", params.UserId)))
			return
		}

		chirp, err := apiCfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
			Body: removeProfanities(params.Body),
			UserID: uuid.NullUUID{
				UUID:  params.UserId,
				Valid: true,
			},
		})
		if err != nil {
			fmt.Printf("Error creating chirp: %v\n", err)
			wrt.WriteHeader(500)
			return
		}

		message := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID.UUID,
		}
		dat, err := json.Marshal(message)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		wrt.WriteHeader(201)
		wrt.Write(dat)
	})

	mux.HandleFunc("GET /api/chirps", func(wrt http.ResponseWriter, req *http.Request) {
		wrt.Header().Set("Content-Type", "application/json")
		chirps, err := apiCfg.db.GetAllChirps(req.Context())
		if err != nil {
			fmt.Printf("Error getting all chirps from DB: %v\n", err)
			wrt.WriteHeader(500)
			return
		}

		messages := make([]Chirp, len(chirps))
		for i, chirp := range chirps {
			messages[i] = Chirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserId:    chirp.UserID.UUID,
			}
		}

		dat, err := json.Marshal(messages)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		wrt.WriteHeader(200)
		wrt.Write(dat)
	})

	mux.HandleFunc("GET /api/chirps/{chirpID}", func(wrt http.ResponseWriter, req *http.Request) {
		wrt.Header().Set("Content-Type", "application/json")
		chirpID := req.PathValue("chirpID")
		if chirpID == "" {
			fmt.Println("failed to parse requested chirp ID")
			wrt.WriteHeader(500)
			return
		}

		id, err := uuid.Parse(chirpID)
		if err != nil {
			wrt.WriteHeader(500)
			fmt.Fprintf(wrt, "Could not parse %v into a uuid", chirpID)
			return
		}
		chirp, err := apiCfg.db.GetChirp(req.Context(), id)
		if err != nil {
			wrt.WriteHeader(404)
			fmt.Fprintf(wrt, "No chirp found for id %v", chirpID)
			return
		}

		message := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID.UUID,
		}

		dat, err := json.Marshal(message)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s\n", err)
			wrt.WriteHeader(500)
			return
		}

		wrt.WriteHeader(200)
		wrt.Write(dat)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server failure.  Error: %v", err)
		os.Exit(1)
	}
}

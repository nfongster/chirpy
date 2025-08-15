package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nfongster/chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := auth.HashPassword(password)
	if hashedPassword == "" || err != nil {
		t.Errorf("Failed to hash password \"%s\"", password)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)

	if err := auth.CheckPasswordHash(password, hashedPassword); err != nil {
		t.Errorf("Failed to check password hash.  Error: %v", err)
	}
}

func TestCheckPasswordHashWrongPassword(t *testing.T) {
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)

	if err := auth.CheckPasswordHash("password2", hashedPassword); err == nil {
		t.Errorf("Expected an error when checking a password hash against an incorrect source password.")
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	id := uuid.New()
	secret := "my_secret"
	ss, err := auth.MakeJWT(id, secret, 10*time.Second)

	if err != nil {
		t.Errorf("making JWT returned err: %v", err)
	}

	validatedId, err := auth.ValidateJWT(ss, secret)

	if err != nil {
		t.Errorf("validating JWT returned err: %v", err)
	}
	if validatedId != id {
		t.Errorf("JWT was validated, but id did not match original id")
	}
}

func TestGetBearerToken(t *testing.T) {
	tokenString := "iamastring"

	headers := make(http.Header)
	headers.Set("Content-Type", "application-json")
	headers.Set("Authorization", "Bearer "+tokenString)

	bearer, err := auth.GetBearerToken(headers)
	if err != nil {
		t.Errorf("error getting bearer token from HTTP header: %v", err)
	}
	if bearer != tokenString {
		t.Errorf("bearer (%s) != tokenString (%s)", bearer, tokenString)
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := auth.MakeRefreshToken()
	if err != nil {
		t.Errorf("error making refresh token: %v", err)
	}
	if token == "" {
		t.Errorf("token was empty string!")
	}
}

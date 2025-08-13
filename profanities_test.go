package main

import "testing"

func TestCleanChirp(t *testing.T) {
	chirp := "I had something interesting for breakfast"

	cleanChirp := removeProfanities(chirp)

	if cleanChirp != chirp {
		t.Errorf("Original chirp \"%s\" did not match cleaned chirp \"%s\"", chirp, cleanChirp)
	}
}

func TestUncleanChirpOneProfanity(t *testing.T) {
	chirp := "I hear Mastodon is better than Chirpy. sharbert I need to migrate"

	cleanChirp := removeProfanities(chirp)

	expectedCleanChirp := "I hear Mastodon is better than Chirpy. **** I need to migrate"
	if cleanChirp != expectedCleanChirp {
		t.Errorf("Expected \"%s\", but got \"%s\"", expectedCleanChirp, cleanChirp)
	}
}

func TestUncleanChirpTwoProfanities(t *testing.T) {
	chirp := "I really need a kerfuffle to go to bed sooner, Fornax !"

	cleanChirp := removeProfanities(chirp)

	expectedCleanChirp := "I really need a **** to go to bed sooner, **** !"
	if cleanChirp != expectedCleanChirp {
		t.Errorf("Expected \"%s\", but got \"%s\"", expectedCleanChirp, cleanChirp)
	}
}

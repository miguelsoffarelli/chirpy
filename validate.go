package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type parameters struct {
	Body string `json:"body"`
}

type cleanParams struct {
	CleanedBody string `json:"cleaned_body"`
}

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	// decode request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	// length check
	const maxChirpLen = 140
	if len(params.Body) > maxChirpLen {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	// OK response
	respondWithJSON(w, http.StatusOK, clearChirp(params))
}

func clearChirp(p parameters) cleanParams {
	body := strings.Split(p.Body, " ")
	for i, word := range body {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			body[i] = "****"
		}
	}

	return cleanParams{
		CleanedBody: strings.Join(body, " "),
	}
}

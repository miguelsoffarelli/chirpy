package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	// decode request body
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	// length check
	const maxChiprLen = 140
	if len(params.Body) > maxChiprLen {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	// OK response
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	respBody := returnVals{
		Valid: true,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

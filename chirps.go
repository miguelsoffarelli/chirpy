package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/miguelsoffarelli/chirpy/internal/database"
)

type chirpParameters struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	params := chirpParameters{}
	err := decodeJSON(r, &params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	if !validateChirp(&params) {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp", nil)
		return
	}

	userUUID, err := uuid.Parse(params.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: userUUID,
	}

	_, err = cfg.DB.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, params)
}

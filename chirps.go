package main

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/miguelsoffarelli/chirpy/internal/database"
)

type chirpParameters struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
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

	chirp, err := cfg.DB.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(chirps) == 0 {
		respondWithError(w, http.StatusOK, "There is no chirps to show", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

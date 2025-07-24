package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/miguelsoffarelli/chirpy/internal/auth"
	"github.com/miguelsoffarelli/chirpy/internal/database"
)

type chirpParameters struct {
	Body string `json:"body"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication error: couldn't get access token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.SECRET)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication error: invalid or expired token", err)
		return
	}

	params := chirpParameters{}
	if err = decodeJSON(r, &params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	if !validateChirp(&params) {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp", nil)
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, mapChirp(chirp))
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error
	author := r.URL.Query().Get("author_id")
	var authorID uuid.UUID

	if author != "" {
		authorID, err = uuid.Parse(author)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author id", err)
			return
		}

		chirps, err = cfg.DB.GetChirpsByAuthor(r.Context(), authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		if len(chirps) == 0 {
			respondWithJSON(w, http.StatusOK, make([]database.Chirp, 0))
			return
		}
	} else {
		chirps, err = cfg.DB.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}

		if len(chirps) == 0 {
			respondWithJSON(w, http.StatusOK, make([]database.Chirp, 0))
			return
		}
	}

	chirpsArr := make([]Chirp, 0)

	for _, chirp := range chirps {
		chirpsArr = append(chirpsArr, mapChirp(chirp))
	}

	respondWithJSON(w, http.StatusOK, chirpsArr)
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

	respondWithJSON(w, http.StatusOK, mapChirp(chirp))
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication error: couldn't get access token", err)
	}

	userID, err := auth.ValidateJWT(token, cfg.SECRET)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication error: invalid or expired token", err)
		return
	}

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

	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Forbidden: can't delete chirps from other users!", err)
		return
	}

	if err := cfg.DB.DeleteChirp(r.Context(), chirpID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: couldn't delete chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func mapChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
}

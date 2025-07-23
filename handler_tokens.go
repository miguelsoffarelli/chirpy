package main

import (
	"net/http"

	"github.com/miguelsoffarelli/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type tokenParams struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization error: couldn't get bearer token", err)
		return
	}

	userID, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization error: Unauthorized", err)
		return
	}

	token, err := auth.MakeJWT(userID, cfg.SECRET)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error: couldn't create access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, tokenParams{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authorization error: couldn't get bearer token", err)
		return
	}

	if err := cfg.DB.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: couldn't revoke refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

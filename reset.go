package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.PLATFORM != "dev" {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN: must be dev to perform this action", nil)
		return
	}

	if err := cfg.DB.ResetUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
	}

	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

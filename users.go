package main

import (
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	type userParameters struct {
		Email string `json:"email"`
	}

	params := userParameters{}
	err := decodeJSON(r, &params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	// Check if email is valid
	if !govalidator.IsEmail(params.Email) {
		respondWithError(w, http.StatusBadRequest, "Email not valid", nil)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if isUniqueConstraintError(err) { // Check for duplicates
		respondWithError(w, http.StatusConflict, "Email already in use, try a different one", nil)
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

// Check for SQL State 23505 for duplicate unique key
func isUniqueConstraintError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

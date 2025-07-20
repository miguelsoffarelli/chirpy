package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/miguelsoffarelli/chirpy/internal/database"
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

	decoder := json.NewDecoder(r.Body)
	params := userParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	// Check if email is valid
	if !govalidator.IsEmail(params.Email) {
		respondWithError(w, http.StatusBadRequest, "Email not valid", nil)
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if isUniqueConstraintError(err) { // Check for duplicates
		respondWithError(w, http.StatusConflict, "Email already in use, try a different one", nil)
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
	}

	respondWithJSON(w, http.StatusCreated, mapUser(user))
}

// Check for SQL State 23505 for duplicate email
func isUniqueConstraintError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

// Helper function to map a database.User into a main.User (the reason behind this
// is to be able to control the JSON keys)
func mapUser(user database.User) User {
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}

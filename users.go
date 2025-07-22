package main

import (
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/miguelsoffarelli/chirpy/internal/auth"
	"github.com/miguelsoffarelli/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	type userParameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	hashedPswd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Password not valid", err)
		return
	}

	createUserParams := database.CreateUserParams{
		HashedPassword: hashedPswd,
		Email:          params.Email,
	}

	user, err := cfg.DB.CreateUser(r.Context(), createUserParams)
	if isUniqueConstraintError(err) { // Check for duplicates
		respondWithError(w, http.StatusConflict, "Email already in use, try a different one", nil)
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, mapUser(user))
}

// Check for SQL State 23505 for duplicate unique key
func isUniqueConstraintError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

func mapUser(user database.User, token ...string) User {
	var userToken string
	if len(token) > 0 {
		userToken = token[0]
	}
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     userToken,
	}
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type loginParams struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	params := loginParams{}
	if err := decodeJSON(r, &params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = 3600
	}

	params.ExpiresInSeconds = min(params.ExpiresInSeconds, 3600)

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	if err := auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.SECRET, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error: Unauthorized", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUser(user, token))
}

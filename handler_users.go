package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/miguelsoffarelli/chirpy/internal/auth"
	"github.com/miguelsoffarelli/chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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

func mapUser(user database.User, tokens ...string) User {
	var userToken string
	var refresh_token string
	if len(tokens) > 0 {
		userToken = tokens[0]
		refresh_token = tokens[1]
	}
	return User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        userToken,
		RefreshToken: refresh_token,
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

	token, err := auth.MakeJWT(user.ID, cfg.SECRET)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication error", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token", err)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		RevokedAt: sql.NullTime{},
	}

	if _, err := cfg.DB.CreateRefreshToken(r.Context(), refreshTokenParams); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error storing refresh token in database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUser(user, token, refreshToken))
}

func (cfg *apiConfig) handlerCredentials(w http.ResponseWriter, r *http.Request) {
	type credentialsParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := credentialsParams{}
	if err := decodeJSON(r, &params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication error: couldn't get access token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.SECRET)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired access token", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error: couldn't hash password", err)
		return
	}

	updateCredentialsParams := database.UpdateCredentialsParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	updatedUser, err := cfg.DB.UpdateCredentials(r.Context(), updateCredentialsParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error: failed to update credentials", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUser(updatedUser))
}

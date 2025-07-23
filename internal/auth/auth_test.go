package auth

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()

	t.Run("basic use case", func(t *testing.T) {
		token, err := MakeJWT(userID, "kerfuffle")
		if err != nil {
			t.Fatalf("error creating token: %v", err)
		}

		id, err := ValidateJWT(token, "kerfuffle")
		if err != nil {
			t.Fatalf("error validating token: %v", err)
		}

		if id != userID {
			t.Fatalf("expected user id %v, got %v", userID, id)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := MakeJWT(userID, "kerfuffle")
		if err != nil {
			t.Fatalf("error creating token: %v", err)
		}

		_, err = ValidateJWT(token, "kerfuffle")
		if err == nil {
			t.Fatalf("no errors returned despite expired token")
		}
	})

	t.Run("create and validate with empty secret", func(t *testing.T) {
		token, err := MakeJWT(userID, "")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		id, err := ValidateJWT(token, "")
		if err != nil {
			t.Fatalf("expected valid JWT with empty secret, got error: %v", err)
		}

		if id != userID {
			t.Fatalf("expected user ID %v, got %v", userID, id)
		}
	})

	t.Run("create with secret, validate with empty secret", func(t *testing.T) {
		token, err := MakeJWT(userID, "kerfuffle")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		_, err = ValidateJWT(token, "")
		if err == nil {
			t.Fatalf("expected error when validating with wrong (empty) secret, got none")
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	headers := make(http.Header)

	t.Run("basic get bearer token case", func(t *testing.T) {
		tokenString := "TOKEN_STRING"
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		token, err := GetBearerToken(headers)
		if err != nil {
			t.Fatalf("error getting token: %v", err)
		}

		if token != "TOKEN_STRING" {
			t.Fatalf("error: expected token %s, got %s", tokenString, token)
		}
	})

	t.Run("get bearer token with empty token", func(t *testing.T) {
		tokenString := ""
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		_, err := GetBearerToken(headers)
		if err == nil {
			t.Fatalf("empty token should return not nil error")
		}
	})

	t.Run("get bearer token with no Authorization header", func(t *testing.T) {
		headers.Del("Authorization")
		_, err := GetBearerToken(headers)
		if err == nil {
			t.Fatalf("absence of Authorization header should return not nil error")
		}
	})

	t.Run("get bearer token with token with whitespaces", func(t *testing.T) {
		tokenString := "token with whitespaces"
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
		_, err := GetBearerToken(headers)
		if err == nil {
			t.Fatalf("expecting error: invalid token, got: nil")
		}
	})
}

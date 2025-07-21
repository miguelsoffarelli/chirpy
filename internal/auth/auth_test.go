package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()

	t.Run("basic use case", func(t *testing.T) {
		token, err := MakeJWT(userID, "kerfuffle", time.Minute)
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
		token, err := MakeJWT(userID, "kerfuffle", -time.Second)
		if err != nil {
			t.Fatalf("error creating token: %v", err)
		}

		_, err = ValidateJWT(token, "kerfuffle")
		if err == nil {
			t.Fatalf("no errors returned despite expired token")
		}
	})

	t.Run("create and validate with empty secret", func(t *testing.T) {
		token, err := MakeJWT(userID, "", time.Minute)
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
		token, err := MakeJWT(userID, "kerfuffle", time.Minute)
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		_, err = ValidateJWT(token, "")
		if err == nil {
			t.Fatalf("expected error when validating with wrong (empty) secret, got none")
		}
	})
}

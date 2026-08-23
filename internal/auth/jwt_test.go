package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}
	if gotID != userID {
		t.Errorf("got userID %v, want %v", gotID, userID)
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	if _, err := ValidateJWT(token, "wrong-secret"); err != ErrInvalidToken {
		t.Errorf("got err %v, want %v", err, ErrInvalidToken)
	}
}

func TestValidateJWTExpired(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "test-secret", -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	if _, err := ValidateJWT(token, "test-secret"); err != ErrInvalidToken {
		t.Errorf("got err %v, want %v", err, ErrInvalidToken)
	}
}

func TestValidateJWTMalformed(t *testing.T) {
	if _, err := ValidateJWT("not-a-real-token", "test-secret"); err != ErrInvalidToken {
		t.Errorf("got err %v, want %v", err, ErrInvalidToken)
	}
}

package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if err := CheckPassword(hash, "correct-horse-battery-staple"); err != nil {
		t.Errorf("CheckPassword returned error for correct password: %v", err)
	}
}

func TestCheckPasswordWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if err := CheckPassword(hash, "wrong-password"); err == nil {
		t.Error("CheckPassword returned nil error for wrong password, want error")
	}
}

func TestHashPasswordUniqueSalt(t *testing.T) {
	hash1, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	hash2, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword produced identical hashes for the same password, want distinct salts")
	}
}

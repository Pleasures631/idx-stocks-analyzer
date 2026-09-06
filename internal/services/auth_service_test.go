package services

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPasswordUsesCost12(t *testing.T) {
	hash, err := HashPassword("securepass123")
	if err != nil {
		t.Fatal(err)
	}
	if cost, err := bcrypt.Cost([]byte(hash)); err != nil || cost != 12 {
		t.Fatalf("expected bcrypt cost 12, got %d (%v)", cost, err)
	}
	if strings.Contains(hash, "securepass123") {
		t.Fatal("password was included in hash output")
	}
}

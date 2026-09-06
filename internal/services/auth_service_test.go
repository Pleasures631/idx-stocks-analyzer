package services

import (
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenerateOTPIsSixDigits(t *testing.T) {
	otp, err := GenerateOTP()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^\d{6}$`).MatchString(otp) {
		t.Fatalf("expected six digit OTP, got %q", otp)
	}
	if !VerifyOTP(HashOTP(otp), otp) {
		t.Fatal("generated OTP did not verify")
	}
	if VerifyOTP(HashOTP(otp), "000000") && otp != "000000" {
		t.Fatal("incorrect OTP verified")
	}
}

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

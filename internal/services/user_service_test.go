package services

import (
	"testing"

	"indonesia-stocks-api/internal/models"
)

func TestNormalizeAndValidateUser(t *testing.T) {
	valid, validationErrors := NormalizeAndValidateUser(models.RegisterUserRequest{
		Name:    "  Siti  Aminah ",
		Phone:   "081234567890",
		Email:   "SITI@example.com",
		Address: "Jl. Merdeka No. 1",
	})
	if len(validationErrors) != 0 {
		t.Fatalf("valid user returned errors: %#v", validationErrors)
	}
	if valid.Name != "Siti Aminah" || valid.Phone != "+6281234567890" || valid.Email != "siti@example.com" {
		t.Fatalf("unexpected normalization: %#v", valid)
	}
}

func TestNormalizeAndValidateUserRejectsInvalidFields(t *testing.T) {
	_, validationErrors := NormalizeAndValidateUser(models.RegisterUserRequest{
		Name:    "A",
		Phone:   "12345",
		Email:   "not-an-email",
		Address: "x",
	})
	for _, field := range []string{"name", "phone", "email", "address"} {
		if _, ok := validationErrors[field]; !ok {
			t.Errorf("expected validation error for %s", field)
		}
	}
}

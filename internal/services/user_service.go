package services

import (
	"net/mail"
	"regexp"
	"strings"

	"indonesia-stocks-api/internal/models"
)

var userNamePattern = regexp.MustCompile(`^[\p{L}][\p{L}\s.'-]{1,119}$`)
var userPhonePattern = regexp.MustCompile(`^(?:\+62|62|0)8[0-9]{8,11}$`)

// NormalizeAndValidateUser normalizes accepted Indonesian phone formats and
// returns field-specific validation errors without touching the database.
func NormalizeAndValidateUser(input models.RegisterUserRequest) (models.RegisterUserRequest, map[string]string) {
	input.Name = strings.Join(strings.Fields(strings.TrimSpace(input.Name)), " ")
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Address = strings.TrimSpace(input.Address)
	errors := make(map[string]string)

	if !userNamePattern.MatchString(input.Name) {
		errors["name"] = "name must contain 2-120 letters and common name characters only"
	}
	if !userPhonePattern.MatchString(input.Phone) {
		errors["phone"] = "phone must use Indonesian format 08..., 62..., or +62..."
	} else {
		input.Phone = normalizeIndonesianPhone(input.Phone)
	}
	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil || parsedEmail.Address != input.Email || len(input.Email) > 254 {
		errors["email"] = "email format is invalid"
	}
	if len([]rune(input.Address)) < 5 || len([]rune(input.Address)) > 500 {
		errors["address"] = "address must contain 5-500 characters"
	}
	return input, errors
}

func normalizeIndonesianPhone(phone string) string {
	if strings.HasPrefix(phone, "+62") {
		return phone
	}
	if strings.HasPrefix(phone, "62") {
		return "+" + phone
	}
	return "+62" + strings.TrimPrefix(phone, "0")
}

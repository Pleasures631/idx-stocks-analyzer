package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
)

const passwordCost = 12

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	return string(hash), err
}

func GenerateOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func HashOTP(otp string) string { return tokenHash(otp) }

func VerifyOTP(expectedHash, otp string) bool {
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(HashOTP(otp))) == 1
}

func Authenticate(email, password string) (*models.AuthResponse, error) {
	user, hash, err := repositories.FindUserByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return issueTokens(user)
}

func issueTokens(user *models.User) (*models.AuthResponse, error) {
	access, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken(48)
	if err != nil {
		return nil, err
	}
	accessExpiry := time.Now().UTC().Add(15 * time.Minute)
	refreshExpiry := time.Now().UTC().Add(30 * 24 * time.Hour)
	if err := repositories.CreateSession(user.ID, tokenHash(access), tokenHash(refresh), accessExpiry, refreshExpiry); err != nil {
		return nil, err
	}
	return &models.AuthResponse{AccessToken: access, RefreshToken: refresh, User: user}, nil
}

func Refresh(refreshToken string) (*models.AuthResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("invalid refresh token")
	}
	access, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken(48)
	if err != nil {
		return nil, err
	}
	user, err := repositories.RotateSession(tokenHash(refreshToken), tokenHash(access), tokenHash(refresh), time.Now().UTC().Add(15*time.Minute), time.Now().UTC().Add(30*24*time.Hour))
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	return &models.AuthResponse{AccessToken: access, RefreshToken: refresh, User: user}, nil
}

func ForgotPassword(email string) error {
	user, _, err := repositories.FindUserByEmail(strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	rawToken, err := randomToken(32)
	if err != nil {
		return err
	}
	if err := repositories.CreatePasswordResetToken(user.ID, tokenHash(rawToken), time.Now().UTC().Add(30*time.Minute)); err != nil {
		return err
	}
	return sendResetEmail(user.Email, user.Name, rawToken)
}

func ResetPassword(token, password string) (*models.User, error) {
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return repositories.ResetPassword(tokenHash(token), string(hash))
}

func SendRegistrationOTP(email, name, otp string) error {
	content := fmt.Sprintf("<h1 style=\"margin:0 0 16px;color:#111827;font-size:24px\">Verify your email</h1><p style=\"color:#4b5563\">Enter this code to complete your StockDash registration:</p><div style=\"margin:24px 0;padding:16px;text-align:center;background:#eef2ff;border-radius:12px;color:#4338ca;font-size:32px;font-weight:700;letter-spacing:8px\">%s</div><p style=\"color:#6b7280;font-size:13px\">This code expires in 10 minutes.</p>", html.EscapeString(otp))
	return sendBrevoEmail(email, name, "Verify your StockDash account", content)
}

func randomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func sendResetEmail(email, name, token string) error {
	baseURL := strings.TrimRight(os.Getenv("FRONTEND_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	content := fmt.Sprintf("<h1 style=\"margin:0 0 16px;color:#111827;font-size:24px\">Reset your password</h1><p style=\"color:#4b5563\">Hi %s, use the button below to choose a new password.</p><p style=\"margin:24px 0\"><a href=\"%s/reset-password?token=%s\" style=\"display:inline-block;padding:12px 18px;border-radius:10px;background:#4f46e5;color:#fff;text-decoration:none;font-weight:600\">Reset password</a></p><p style=\"color:#6b7280;font-size:13px\">This link expires in 30 minutes. If you did not request it, you can ignore this email.</p>", html.EscapeString(name), html.EscapeString(baseURL), html.EscapeString(token))
	return sendBrevoEmail(email, name, "Reset your StockDash password", content)
}

func sendBrevoEmail(email, name, subject, content string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := os.Getenv("BREVO_SENDER_EMAIL")
	if apiKey == "" || senderEmail == "" {
		return errors.New("Brevo email configuration is missing")
	}
	logoURL := strings.TrimRight(os.Getenv("BRAND_LOGO_URL"), "/")
	if logoURL == "" {
		logoURL = strings.TrimRight(os.Getenv("FRONTEND_BASE_URL"), "/") + "/img/yapping-saham-logo/logo.png"
	}
	payload := map[string]any{
		"sender":      map[string]string{"name": os.Getenv("BREVO_SENDER_NAME"), "email": senderEmail},
		"to":          []map[string]string{{"email": email, "name": name}},
		"subject":     subject,
		"htmlContent": fmt.Sprintf("<div style=\"background:#f3f4f6;padding:32px 16px;font-family:Arial,sans-serif\"><div style=\"max-width:520px;margin:auto;background:#fff;border-radius:16px;padding:32px;box-shadow:0 4px 18px rgba(15,23,42,.08)\"><img src=\"%s\" alt=\"Yapping Saham\" width=240 style=\"display:block;width:240px;max-width:100%%;height:auto;margin:0 0 28px\"><div>%s</div></div></div>", html.EscapeString(logoURL), content),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		LogInterfaceCall("SendPasswordResetEmail", req.URL.String(), "", 0, err)
		return err
	}
	defer resp.Body.Close()
	var responseBody bytes.Buffer
	_, _ = responseBody.ReadFrom(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("Brevo returned HTTP %d", resp.StatusCode)
	}
	LogInterfaceCall("SendBrevoEmail", req.URL.String(), responseBody.String(), resp.StatusCode, err)
	return err
}

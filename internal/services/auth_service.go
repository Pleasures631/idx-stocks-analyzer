package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := os.Getenv("BREVO_SENDER_EMAIL")
	if apiKey == "" || senderEmail == "" {
		return errors.New("Brevo email configuration is missing")
	}
	baseURL := strings.TrimRight(os.Getenv("FRONTEND_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	payload := map[string]any{
		"sender":      map[string]string{"name": os.Getenv("BREVO_SENDER_NAME"), "email": senderEmail},
		"to":          []map[string]string{{"email": email, "name": name}},
		"subject":     "Reset your StockDash password",
		"htmlContent": fmt.Sprintf("<p>Use this link to reset your password:</p><p><a href=\"%s/reset-password?token=%s\">Reset password</a></p><p>This link expires in 30 minutes.</p>", baseURL, token),
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
	LogInterfaceCall("SendPasswordResetEmail", req.URL.String(), responseBody.String(), resp.StatusCode, err)
	return err
}

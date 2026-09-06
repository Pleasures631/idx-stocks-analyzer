package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterUser(c *gin.Context) {
	var request models.RegisterUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid JSON body"})
		return
	}

	normalized, validationErrors := services.NormalizeAndValidateUser(request)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":           false,
			"message":           "validation failed",
			"validation_errors": validationErrors,
		})
		return
	}
	alreadyRegistered, err := repositories.UserAlreadyRegistered(normalized.Email, normalized.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to check user registration"})
		return
	}
	if alreadyRegistered {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "email or phone already registered"})
		return
	}
	passwordHash, err := services.HashPassword(normalized.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to secure password"})
		return
	}
	normalized.Password = string(passwordHash)

	otp, err := services.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to create verification code"})
		return
	}
	if err := repositories.CreatePendingRegistration(normalized, services.HashOTP(otp), time.Now().UTC().Add(10*time.Minute)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to start registration"})
		return
	}
	if err := services.SendRegistrationOTP(normalized.Email, normalized.Name, otp); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "verification email could not be sent"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true, "message": "verification code sent to email"})
}

func VerifyRegistration(c *gin.Context) {
	var request models.VerifyRegistrationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid JSON body"})
		return
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if len(request.OTP) != 6 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "verification code is invalid"})
		return
	}
	pending, err := repositories.FindPendingRegistration(request.Email)
	if err != nil || errors.Is(err, sql.ErrNoRows) || pending == nil || pending.Attempts >= 5 || time.Now().UTC().After(pending.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "verification code is invalid or expired"})
		return
	}
	if !services.VerifyOTP(pending.OTPHash, request.OTP) {
		_ = repositories.IncrementPendingRegistrationAttempts(pending.ID)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "verification code is invalid"})
		return
	}
	user, err := repositories.CreateUser(models.RegisterUserRequest{Name: pending.Name, Phone: pending.Phone, Email: pending.Email, Address: pending.Address, Password: pending.PasswordHash})
	if err != nil {
		if errors.Is(err, repositories.ErrUserConflict) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "email or phone already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to register user"})
		return
	}
	if err := repositories.DeletePendingRegistration(pending.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "user created but verification cleanup failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": user})
}

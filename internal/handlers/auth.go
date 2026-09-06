package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/services"
)

func Login(c *gin.Context) {
	var req models.LoginRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid JSON body"})
		return
	}
	result, err := services.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid email or password"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid JSON body"})
		return
	}
	result, err := services.Refresh(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid refresh token"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid JSON body"})
		return
	}
	// Keep account existence private. Delivery/configuration failures are logged
	// server-side and do not disclose whether the address is registered.
	_ = services.ForgotPassword(req.Email)
	c.JSON(http.StatusAccepted, gin.H{"success": true, "message": "if the email is registered, a reset link will be sent"})
}

func ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if c.ShouldBindJSON(&req) != nil || req.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "token and password are required"})
		return
	}
	user, err := services.ResetPassword(req.Token, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid or expired reset token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

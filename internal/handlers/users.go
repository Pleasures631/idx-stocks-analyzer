package handlers

import (
	"errors"
	"net/http"

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

	user, err := repositories.CreateUser(normalized)
	if err != nil {
		if errors.Is(err, repositories.ErrUserConflict) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "email or phone already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": user})
}

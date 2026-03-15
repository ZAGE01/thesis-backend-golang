package handlers

import (
	"game-backend/database"
	"game-backend/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to process password"})
		return
	}

	var userID uint
	err = database.DB.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		req.Username, string(hashedPassword),
	).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "error": "Username already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Account created successfully",
		"user_id": userID,
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": err.Error()})
		return
	}

	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, password, is_admin FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid credentials"})
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"token":    tokenStr,
		"user_id":  user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
	})
}

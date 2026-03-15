package handlers

import (
	"game-backend/database"
	"game-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SubmitScore(c *gin.Context) {
	var req models.ScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Value < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Score cannot be negative"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	var scoreID uint
	err := database.DB.QueryRow(
		"INSERT INTO scores (user_id, value) VALUES ($1, $2) RETURNING id",
		userID, req.Value,
	).Scan(&scoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit score"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Score submitted successfully",
		"score_id": scoreID,
	})
}

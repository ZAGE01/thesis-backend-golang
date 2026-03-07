package handlers

import (
	"game-backend/database"
	"game-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLeaderboard(c *gin.Context) {
	rows, err := database.DB.Query(`
        SELECT u.username, MAX(s.value) as top_score
        FROM scores s
        JOIN users u ON s.user_id = u.id
        GROUP BY u.username
        ORDER BY top_score DESC
        LIMIT 100
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry models.LeaderboardEntry
		if err := rows.Scan(&entry.Username, &entry.Score); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse leaderboard"})
			return
		}
		entry.Rank = rank
		entries = append(entries, entry)
		rank++
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"total":       len(entries),
	})
}

func GetPlayer(c *gin.Context) {
	playerID := c.Param("id")

	var username string
	err := database.DB.QueryRow(
		"SELECT username FROM users WHERE id = $1", playerID,
	).Scan(&username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Get top score and total games played
	var topScore, gamesPlayed int
	database.DB.QueryRow(`
        SELECT COALESCE(MAX(value), 0), COUNT(*) 
        FROM scores WHERE user_id = $1`, playerID,
	).Scan(&topScore, &gamesPlayed)

	c.JSON(http.StatusOK, gin.H{
		"username":     username,
		"top_score":    topScore,
		"games_played": gamesPlayed,
	})
}

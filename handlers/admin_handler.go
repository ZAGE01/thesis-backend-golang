package handlers

import (
	"game-backend/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListUsers(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT id, username, is_admin, created_at FROM users ORDER BY id ASC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []gin.H
	for rows.Next() {
		var id int
		var username string
		var isAdmin bool
		var createdAt string
		if err := rows.Scan(&id, &username, &isAdmin, &createdAt); err != nil {
			continue
		}
		users = append(users, gin.H{
			"id":         id,
			"username":   username,
			"is_admin":   isAdmin,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "total": len(users)})
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result, err := database.DB.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

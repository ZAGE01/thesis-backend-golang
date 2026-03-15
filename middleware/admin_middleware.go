package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || isAdmin != true {
			c.JSON(http.StatusForbidden, gin.H{"status": http.StatusForbidden, "error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

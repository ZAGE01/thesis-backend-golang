package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// readJSON is a shared helper for both middleware test files
func readJSON(w *httptest.ResponseRecorder, v interface{}) {
	json.Unmarshal(w.Body.Bytes(), v)
}

// Helper to make a request with is_admin pre-set in context
func makeAdminRequest(isAdmin interface{}, exists bool) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		if exists {
			c.Set("is_admin", isAdmin)
		}
		AdminMiddleware()(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)
	return w
}

func TestAdminMiddleware_AdminUser_Passes(t *testing.T) {
	w := makeAdminRequest(true, true)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddleware_NonAdminUser_Forbidden(t *testing.T) {
	w := makeAdminRequest(false, true)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Admin access required", resp["error"])
}

func TestAdminMiddleware_MissingIsAdminClaim_Forbidden(t *testing.T) {
	// is_admin not set in context at all (exists = false)
	w := makeAdminRequest(nil, false)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Admin access required", resp["error"])
}

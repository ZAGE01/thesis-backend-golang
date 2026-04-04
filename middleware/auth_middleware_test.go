package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func setupAuthMiddlewareRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		isAdmin, _ := c.Get("is_admin")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"is_admin": isAdmin,
		})
	})
	return r
}

func generateTestToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	// No Authorization header
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Authorization header required", resp["error"])
}

func TestAuthMiddleware_MalformedHeader_NoBearer(t *testing.T) {
	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "sometoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Authorization header format must be: Bearer <token>", resp["error"])
}

func TestAuthMiddleware_MalformedHeader_WrongScheme(t *testing.T) {
	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Authorization header format must be: Bearer <token>", resp["error"])
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer notavalidtoken")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Invalid or expired token", resp["error"])
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	tokenStr := generateTestToken("test-secret", jwt.MapClaims{
		"user_id":  float64(1),
		"username": "user1",
		"is_admin": false,
		"exp":      time.Now().Add(-1 * time.Hour).Unix(), // already expired
	})

	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Invalid or expired token", resp["error"])
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	// Token signed with a different secret
	tokenStr := generateTestToken("wrong-secret", jwt.MapClaims{
		"user_id":  float64(1),
		"username": "user1",
		"is_admin": false,
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})

	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, "Invalid or expired token", resp["error"])
}

func TestAuthMiddleware_ValidToken_ClaimsSetInContext(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	tokenStr := generateTestToken("test-secret", jwt.MapClaims{
		"user_id":  float64(1),
		"username": "user1",
		"is_admin": false,
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})

	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, float64(1), resp["user_id"])
	assert.Equal(t, "user1", resp["username"])
	assert.Equal(t, false, resp["is_admin"])
}

func TestAuthMiddleware_ValidToken_AdminClaims(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	tokenStr := generateTestToken("test-secret", jwt.MapClaims{
		"user_id":  float64(2),
		"username": "adminuser",
		"is_admin": true,
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})

	r := setupAuthMiddlewareRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	readJSON(w, &resp)
	assert.Equal(t, true, resp["is_admin"])
	assert.Equal(t, "adminuser", resp["username"])
}

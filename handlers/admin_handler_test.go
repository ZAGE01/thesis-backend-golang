package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"game-backend/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAdminRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/users", ListUsers)
	r.DELETE("/admin/users/:id", DeleteUser)
	return r
}

// List users endpoint

func TestListUsers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	createdAt := time.Now().Format(time.RFC3339)
	rows := sqlmock.NewRows([]string{"id", "username", "is_admin", "created_at"}).
		AddRow(1, "user1", false, createdAt).
		AddRow(2, "user2", true, createdAt)

	mock.ExpectQuery(`SELECT id, username, is_admin, created_at FROM users ORDER BY id ASC`).
		WillReturnRows(rows)

	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, float64(http.StatusOK), resp["status"])
	assert.Equal(t, float64(2), resp["total"])

	users := resp["users"].([]interface{})
	assert.Len(t, users, 2)
	assert.Equal(t, "user1", users[0].(map[string]interface{})["username"])
	assert.Equal(t, "user2", users[1].(map[string]interface{})["username"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListUsers_EmptyDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "is_admin", "created_at"})
	mock.ExpectQuery(`SELECT id, username, is_admin, created_at FROM users ORDER BY id ASC`).
		WillReturnRows(rows)

	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, float64(0), resp["total"])
	assert.Nil(t, resp["users"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListUsers_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`SELECT id, username, is_admin, created_at FROM users ORDER BY id ASC`).
		WillReturnError(fmt.Errorf("connection lost"))

	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to fetch users", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Delete users endpoint

func TestDeleteUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "User deleted successfully", resp["message"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/admin/users/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "User not found", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/admin/users/invalid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid user ID", resp["error"])
}

func TestDeleteUser_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnError(fmt.Errorf("connection lost"))

	r := setupAdminRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to delete user", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

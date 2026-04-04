package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"game-backend/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", Register)
	r.POST("/login", Login)
	return r
}

// Register endpoint

func TestRegister_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("testuser", sqlmock.AnyArg()). // AnyArg() for hashed password
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "password123",
	})

	r := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Account created successfully", resp["message"])
	assert.Equal(t, float64(1), resp["user_id"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_DuplicateUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("testuser", sqlmock.AnyArg()). // AnyArg() for hashed password
		WillReturnError(fmt.Errorf("unique constraint violation"))

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "password123",
	})

	r := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Username already taken", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_MissingUsername(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]string{
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Username cannot be empty", resp["error"])
}

func TestRegister_UsernameTooShort(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]string{
		"username": "te",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Username must be at least 3 characters", resp["error"])
}

func TestRegister_UsernameTooLong(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]string{
		"username": "thisusernameiswaytoolong123",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Username must be less than 20 characters", resp["error"])
}

func TestRegister_UsernameWithSpaces(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]string{
		"username": "test user",
		"password": "password123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Username cannot contain spaces", resp["error"])
}

func TestRegister_PasswordTooShort(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Password must be at least 6 characters", resp["error"])
}

func TestRegister_InvalidBody(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid request body", resp["error"])
}

// Login endpoint

func TestLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	os.Setenv("JWT_SECRET", "test-secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{"id", "username", "password", "is_admin"}).
		AddRow(1, "testuser", string(hashed), false)

	mock.ExpectQuery(`SELECT id, username, password, is_admin FROM users WHERE username`).
		WithArgs("testuser").
		WillReturnRows(rows)

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "password123",
	})

	r := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])
	assert.Equal(t, "testuser", resp["username"])
	assert.Equal(t, float64(1), resp["user_id"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`SELECT id, username, password, is_admin FROM users WHERE username`).
		WithArgs("unknown").
		WillReturnError(fmt.Errorf("no rows"))

	body, _ := json.Marshal(map[string]string{
		"username": "unknown",
		"password": "password123",
	})

	r := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid credentials", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_WrongPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{"id", "username", "password", "is_admin"}).
		AddRow(1, "testuser", string(hashed), false)

	mock.ExpectQuery(`SELECT id, username, password, is_admin FROM users WHERE username`).
		WithArgs("testuser").
		WillReturnRows(rows)

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	})

	r := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid credentials", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_MissingFields(t *testing.T) {
	r := setupAuthRouter()
	w := httptest.NewRecorder()

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		// missing password
	})

	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Password cannot be empty", resp["error"])
}

func TestLogin_AdminUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	os.Setenv("JWT_SECRET", "test-secret")

	hashed, _ := bcrypt.GenerateFromPassword([]byte("adminpass"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{"id", "username", "password", "is_admin"}).
		AddRow(2, "adminuser", string(hashed), true) // is_admin = true

	mock.ExpectQuery(`SELECT id, username, password, is_admin FROM users WHERE username`).
		WithArgs("adminuser").
		WillReturnRows(rows)

	body, _ := json.Marshal(map[string]string{
		"username": "adminuser",
		"password": "adminpass",
	})

	r := setupAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["is_admin"])
	assert.NotEmpty(t, resp["token"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"game-backend/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupScoreRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Simulate AuthMiddleware by injecting user_id into context directly
	r.POST("/score", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		SubmitScore(c)
	})
	return r
}

// Score submitting endpoint

func TestSubmitScore_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO scores`).
		WithArgs(uint(1), 100).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	body, _ := json.Marshal(map[string]int{"value": 100})

	r := setupScoreRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/score", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Score submitted successfully", resp["message"])
	assert.Equal(t, float64(1), resp["score_id"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmitScore_NegativeValue(t *testing.T) {
	// No DB interaction expected — handler rejects before querying
	body, _ := json.Marshal(map[string]int{"value": -10})

	r := setupScoreRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/score", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Score cannot be negative", resp["error"])
}

func TestSubmitScore_MissingOrZeroValue(t *testing.T) {
	// ScoreRequest has binding:"required" on Value
	body, _ := json.Marshal(map[string]string{})
	body2, _ := json.Marshal(map[string]int{"value": 0})

	r := setupScoreRouter()
	w := httptest.NewRecorder()

	// Missing "value" field
	req, _ := http.NewRequest(http.MethodPost, "/score", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// "value" field is zero
	req2, _ := http.NewRequest(http.MethodPost, "/score", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req2)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitScore_InvalidBody(t *testing.T) {
	r := setupScoreRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/score", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitScore_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO scores`).
		WithArgs(uint(1), 100).
		WillReturnError(fmt.Errorf("connection lost"))

	body, _ := json.Marshal(map[string]int{"value": 100})

	r := setupScoreRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/score", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to submit score", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

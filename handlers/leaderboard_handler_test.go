package handlers

import (
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

func setupLeaderboardRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/leaderboard", GetLeaderboard)
	r.GET("/player/:id", GetPlayer)
	return r
}

// Leaderboard fetching endpoint

func TestGetLeaderboard_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	rows := sqlmock.NewRows([]string{"username", "top_score"}).
		AddRow("user1", 150).
		AddRow("user2", 100).
		AddRow("user3", 50)

	mock.ExpectQuery(`SELECT u.username, MAX\(s.value\) as top_score`).
		WillReturnRows(rows)

	r := setupLeaderboardRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/leaderboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, float64(http.StatusOK), resp["status"])
	assert.Equal(t, float64(3), resp["total"])

	entries := resp["leaderboard"].([]interface{})
	assert.Len(t, entries, 3)

	// Check ranking order and values
	first := entries[0].(map[string]interface{})
	assert.Equal(t, "user1", first["username"])
	assert.Equal(t, float64(150), first["score"])
	assert.Equal(t, float64(1), first["rank"])

	second := entries[1].(map[string]interface{})
	assert.Equal(t, "user2", second["username"])
	assert.Equal(t, float64(2), second["rank"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLeaderboard_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	rows := sqlmock.NewRows([]string{"username", "top_score"})
	mock.ExpectQuery(`SELECT u.username, MAX\(s.value\) as top_score`).
		WillReturnRows(rows)

	r := setupLeaderboardRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/leaderboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, float64(0), resp["total"])
	assert.Nil(t, resp["leaderboard"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLeaderboard_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`SELECT u.username, MAX\(s.value\) as top_score`).
		WillReturnError(fmt.Errorf("connection lost"))

	r := setupLeaderboardRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/leaderboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to fetch leaderboard", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Get player details endpoint

func TestGetPlayer_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	// First query: fetch username
	mock.ExpectQuery(`SELECT username FROM users WHERE id`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("user1"))

	// Second query: fetch top score + games played
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(value\), 0\), COUNT\(\*\)`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"top_score", "games_played"}).AddRow(200, 5))

	r := setupLeaderboardRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/player/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, float64(http.StatusOK), resp["status"])
	assert.Equal(t, "user1", resp["username"])
	assert.Equal(t, float64(200), resp["top_score"])
	assert.Equal(t, float64(5), resp["games_played"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPlayer_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	mock.ExpectQuery(`SELECT username FROM users WHERE id`).
		WithArgs("999").
		WillReturnError(fmt.Errorf("no rows"))

	r := setupLeaderboardRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/player/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Player not found", resp["error"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPlayer_NoScores(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	database.DB = db
	defer db.Close()

	// Player exists but has never submitted a score
	mock.ExpectQuery(`SELECT username FROM users WHERE id`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("user1"))

	// COALESCE ensures top_score returns 0, COUNT returns 0
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(value\), 0\), COUNT\(\*\)`).
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"top_score", "games_played"}).AddRow(0, 0))

	r := setupLeaderboardRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/player/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "user1", resp["username"])
	assert.Equal(t, float64(0), resp["top_score"])
	assert.Equal(t, float64(0), resp["games_played"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

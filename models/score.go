package models

import "time"

type Score struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Value       int       `json:"value"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type ScoreRequest struct {
	Value int `json:"value" binding:"required"`
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

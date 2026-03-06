# Game Backend API

A REST API backend built with **Go (Gin)** and **PostgreSQL** for an Unreal Engine game. It handles player account registration, authentication, score submission, and leaderboard tracking.

---

## Tech Stack

- **Language:** Go
- **Framework:** Gin
- **Database:** PostgreSQL
- **Auth:** JWT
- **Password Hashing:** bcrypt

---

## Project Structure

```
thesis-backend-golang/
├── database/
│   ├── db.go         # Database connection & table creation
│   └── db.sql        # SQL schema
├── handlers/
│   ├── auth_handler.go        # Register & Login
│   ├── leaderboard_handler.go # Leaderboard & Player profile
│   └── score_handler.go       # Score submission
├── middleware/
│   └── auth_middleware.go     # JWT auth middleware
├── models/
│   ├── user.go   # User models & request types
│   └── score.go  # Score models & request types
├── main.go
├── go.mod
└── .env
```

---

## Setup

### 1. Prerequisites
- Go 1.26.1
- PostgreSQL

### 2. Configure environment

Create a `.env` file in the root directory:

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=
DB_NAME=gamedb
DB_PORT=5432
JWT_SECRET=
PORT=8080
```

### 3. Run the application

```bash
go run .
```

The server will automatically create the `gamedb` database and required tables on first run.

---

## API Endpoints

### Public Routes

| Method | Endpoint    | Description                        |
|--------|-------------|------------------------------------|
| POST   | `/register` | Create a new account               |
| POST   | `/login`    | Login and get JWT token            |
| GET    | `/leaderboard`  | Get top 10 leaderboard         |
| GET    | `/player/:id`   | Get a player's profile & stats |

### Protected Routes (require `Authorization: Bearer <token>` header)

| Method | Endpoint        | Description                    |
|--------|-----------------|--------------------------------|
| POST   | `/score`        | Submit a score                 |

---

## Example curl Requests

### Register a new account

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"test\", \"password\": \"testpwd\"}"
```

**Response:**
```json
{
  "message": "Account created successfully",
  "user_id": 1
}
```

---

### Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"test\", \"password\": \"testpwd\"}"
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": 1,
  "username": "test"
}
```

> Save the `token` value — you'll need it for all protected requests.

---

### Submit a Score

```bash
curl -X POST http://localhost:8080/score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d "{\"value\": 4500}"
```

**Response:**
```json
{
  "message": "Score submitted successfully",
  "score_id": 7
}
```

---

### Get Leaderboard

```bash
curl -X GET http://localhost:8080/leaderboard
```

**Response:**
```json
{
  "leaderboard": [
    { "rank": 1, "username": "test", "score": 9800 },
    { "rank": 2, "username": "player2", "score": 7500 },
    { "rank": 3, "username": "player3", "score": 6200 }
  ],
  "total": 3
}
```

---

### Get Player Profile

```bash
curl -X GET http://localhost:8080/player/1
```

**Response:**
```json
{
  "username": "test",
  "top_score": 9800,
  "games_played": 12
}
```

---

## Database Schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scores (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    value INTEGER NOT NULL,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
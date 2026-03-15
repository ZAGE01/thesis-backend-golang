# Game Backend API

A REST API backend built with **Go (Gin)** and **PostgreSQL** for an Unreal Engine game. It handles player account registration, authentication, score submission, and leaderboard tracking. This project is part of a thesis project.

---

## Tech Stack

- **Language:** Go
- **Framework:** Gin
- **Database:** PostgreSQL database provided by Supabase
- **Deployment:** Deployed to Render at: https://thesis-backend-golang.onrender.com/
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
│   ├── admin_handler.go       # Endpoints for admins
│   ├── auth_handler.go        # Register & Login
│   ├── leaderboard_handler.go # Leaderboard & Player profile
│   └── score_handler.go       # Score submission
├── middleware/
│   ├── admin_middleware.go    # Admin middleware to restrict access some endpoints
│   └── auth_middleware.go     # JWT auth middleware
├── models/
│   ├── user.go   # User models & request types
│   └── score.go  # Score models & request types
├── main.go
├── go.mod
└── .env
```

---

## Running locally for development

### 1. Prerequisites
- Go 1.25.0

### 2. Configure environment

Create a `.env` file in the root directory and set variables to match Supabase connection parameters:

```env
DB_HOST=?
DB_USER=?
DB_PASSWORD=?
DB_NAME=postgres
DB_PORT=5432
JWT_SECRET=?
PORT=8080
```

### 3. Run the application

```bash
go run .
```

---

## API Endpoints

### Public Routes

| Method | Endpoint        | Description                    |
|--------|-----------------|--------------------------------|
| POST   | `/register`     | Create a new account           |
| POST   | `/login`        | Login and get JWT token        |
| GET    | `/leaderboard`  | Get top 100 leaderboard        |
| GET    | `/player/:id`   | Get a player's profile & stats |

### Protected Routes (require `Authorization: Bearer <token>` header)

| Method | Endpoint        | Description                    |
|--------|-----------------|--------------------------------|
| POST   | `/score`        | Submit a score                 |

### Admin Routes (require `Authorization: Bearer <token>` header and admin role)

| Method | Endpoint           | Description                 |
|--------|--------------------|-----------------------------|
| GET    | `/admin/users`     | List all users              |
| DELETE | `/admin/users/:id` | Delete user by ID           |

---

## Example curl Requests

### Register a new account

```bash
curl -X POST https://thesis-backend-golang.onrender.com/register \
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
curl -X POST https://thesis-backend-golang.onrender.com/login \
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

### Get Leaderboard

```bash
curl -X GET https://thesis-backend-golang.onrender.com/leaderboard
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
curl -X GET https://thesis-backend-golang.onrender.com/player/1
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

### Submit a Score (Requires token)

```bash
curl -X POST https://thesis-backend-golang.onrender.com/score \
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

### List All Users (Admin only)

```bash
curl -X GET https://thesis-backend-golang.onrender.com/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Response:**
```json
{
  "users": [
    { "id": 1, "username": "test", "is_admin": true, "created_at": "2026-01-01T00:00:00Z" },
    { "id": 2, "username": "player2", "is_admin": false, "created_at": "2026-01-02T00:00:00Z" }
  ],
  "total": 2
}
```

---

### Delete User (Admin only)

```bash
curl -X DELETE https://thesis-backend-golang.onrender.com/admin/users/2 \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Response:**
```json
{
  "message": "User deleted successfully"
}
```

---

## Database Schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scores (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    value INTEGER NOT NULL,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
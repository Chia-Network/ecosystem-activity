package users

import (
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
)

// User represents all columns in one user entry in the users table
type User struct {
	ID          int
	Username    string
	Email       string
	Name        string
	CreatedAt   time.Time
	FirstCommit time.Time
	LastCommit  time.Time
	Notes       string
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(u User) error {
	_, err := db.Exec(`INSERT INTO users (username,email,name,created_at,first_commit,last_commit,notes) VALUES(?, ?, ?, ?, ?);`, u.Username, u.Email, u.Name, u.CreatedAt, u.FirstCommit, u.LastCommit, u.Notes)
	return err
}

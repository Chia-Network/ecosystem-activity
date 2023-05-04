package users

import (
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
)

// User represents all columns in one user entry in the users table
type User struct {
	ID          int
	Username    string
	FirstCommit time.Time
	LastCommit  time.Time
	Notes       string
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(u User) ([]User, error) {
	var users []User
	rows, err := db.Query(`INSERT INTO users (username,first_commit,last_commit,notes) VALUES(?, ?, ?, ?);`, u.Username, u.FirstCommit, u.LastCommit, u.Notes)
	if err != nil {
		return []User{}, fmt.Errorf("error adding user to users table for \"%s\": %v", u.Username, err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.FirstCommit, &user.LastCommit, &user.FirstCommit, &user.Notes)
		if err != nil {
			return []User{}, fmt.Errorf("error scanning row for user \"%s\": %v", u.Username, err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return []User{}, fmt.Errorf("error encountered iterating through rows for user \"%s\": %v", u.Username, err)
	}

	return users, nil
}

// GetRowsByUsername gets a slice of rows matching a username
func GetRowsByUsername(username string) ([]User, error) {
	var users []User
	rows, err := db.Query("SELECT id,username,first_commit,last_commit,notes FROM users WHERE username = ?", username)
	if err != nil {
		return users, fmt.Errorf("error querying users table for rows by username \"%s\": %v", username, err)
	}
	defer rows.Close()

	for rows.Next() {
		var r User
		err := rows.Scan(&r.ID, &r.Username, &r.FirstCommit, &r.LastCommit, &r.Notes)
		if err != nil {
			return users, fmt.Errorf("error scanning row for username \"%s\": %v", username, err)
		}
		users = append(users, r)
	}
	if err := rows.Err(); err != nil {
		return users, fmt.Errorf("error encountered iterating through rows for username \"%s\": %v", username, err)
	}

	return users, nil
}

// UpdateLastCommitByUsername accepts a username and time object and updates the matching row's last_commit column to the timestamp
func UpdateLastCommitByUsername(username string, ts time.Time) error {
	_, err := db.Exec(`UPDATE users SET last_commit='?' WHERE username='?';`, ts.Format("2006-01-02 15:04:05"), username)
	if err != nil {
		return fmt.Errorf("error encountered updating last_commit on row for %s: %v", username, err)
	}
	return err
}

// UpdateFirstCommitByUsername accepts a username and time object and updates the matching row's first_commit column to the timestamp
func UpdateFirstCommitByUsername(username string, ts time.Time) error {
	_, err := db.Exec(`UPDATE users SET first_commit='?' WHERE username='?';`, ts.Format("2006-01-02 15:04:05"), username)
	if err != nil {
		return fmt.Errorf("error encountered updating first_commit on row for %s: %v", username, err)
	}
	return err
}

package users

import (
	"database/sql"
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

// userWithNulls is a helper struct for mysql rows that may contain null fields
// a null field using the mysql database driver won't scan into the appropriate field
type userWithNulls struct {
	ID          sql.NullInt64
	Username    sql.NullString
	FirstCommit sql.NullTime
	LastCommit  sql.NullTime
	Notes       sql.NullString
}

// convertSqlUserToUser handles the internal conversion between an sql row response and a user-friendly User struct
// because Go's sql package errors when scanning nil columns in a row
func convertSqlUserToUser(u userWithNulls) User {
	var user User
	if u.ID.Valid {
		user.ID = int(u.ID.Int64)
	}
	if u.Username.Valid {
		user.Username = u.Username.String
	}
	if u.FirstCommit.Valid {
		user.FirstCommit = u.FirstCommit.Time
	}
	if u.LastCommit.Valid {
		user.LastCommit = u.LastCommit.Time
	}
	if u.Notes.Valid {
		user.Notes = u.Notes.String
	}
	return user
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
		var uWithNull userWithNulls
		err := rows.Scan(&uWithNull.ID, &uWithNull.Username, &uWithNull.FirstCommit, &uWithNull.LastCommit, &uWithNull.Notes)
		if err != nil {
			return users, fmt.Errorf("error scanning row for username \"%s\": %v", username, err)
		}
		nonNullUser := convertSqlUserToUser(uWithNull)
		users = append(users, nonNullUser)
	}
	if err := rows.Err(); err != nil {
		return users, fmt.Errorf("error encountered iterating through rows for username \"%s\": %v", username, err)
	}

	return users, nil
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(u User) error {
	_, err := db.Exec(`INSERT INTO users (username,first_commit,last_commit,notes) VALUES(?, ?, ?, ?);`, u.Username, u.FirstCommit, u.LastCommit, u.Notes)
	if err != nil {
		return fmt.Errorf("error adding user to users table for \"%s\": %v", u.Username, err)
	}

	return nil
}

// UpdateLastCommitByUsername accepts a username and time object and updates the matching row's last_commit column to the timestamp
func UpdateLastCommitByUsername(username string, ts time.Time) error {
	_, err := db.Exec(`UPDATE users SET last_commit=? WHERE username=?;`, ts.Format("2006-01-02 15:04:05"), username)
	if err != nil {
		return fmt.Errorf("error encountered updating last_commit on row for %s: %v", username, err)
	}
	return err
}

// UpdateFirstCommitByUsername accepts a username and time object and updates the matching row's first_commit column to the timestamp
func UpdateFirstCommitByUsername(username string, ts time.Time) error {
	_, err := db.Exec(`UPDATE users SET first_commit=? WHERE username=?;`, ts.Format("2006-01-02 15:04:05"), username)
	if err != nil {
		return fmt.Errorf("error encountered updating first_commit on row for %s: %v", username, err)
	}
	return err
}

package commits

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
	log "github.com/sirupsen/logrus"
)

// Commit represents all columns in one commit entry in the commits table
type Commit struct {
	ID     int
	RepoID int
	UserID int
	Date   time.Time
	SHA    string
	Notes  string
}

// commitWithNulls is a helper struct for mysql rows that may contain null fields
// a null field using the mysql database driver won't scan into the appropriate field
type commitWithNulls struct {
	ID     sql.NullInt64
	RepoID sql.NullInt64
	UserID sql.NullInt64
	Date   sql.NullTime
	SHA    sql.NullString
	Notes  sql.NullString
}

// convertSQLRepoToRepo handles the internal conversion between an sql row response and a user-friendly commit struct
// because Go's sql package errors when scanning nil columns in a row
func convertSQLCommitToCommit(c commitWithNulls) Commit {
	var commit Commit
	if c.ID.Valid {
		commit.ID = int(c.ID.Int64)
	}
	if c.RepoID.Valid {
		commit.RepoID = int(c.RepoID.Int64)
	}
	if c.UserID.Valid {
		commit.UserID = int(c.UserID.Int64)
	}
	if c.Date.Valid {
		commit.Date = c.Date.Time
	}
	if c.SHA.Valid {
		commit.SHA = c.SHA.String
	}
	if c.Notes.Valid {
		commit.Notes = c.Notes.String
	}
	return commit
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(c Commit) error {
	_, err := db.Exec(`INSERT INTO commits (repo_id,user_id,date,sha,notes) VALUES(?, ?, ?, ?, ?);`, c.RepoID, c.UserID, c.Date.Format("2006-01-02 15:04:05"), c.SHA, c.Notes)
	if err != nil {
		return fmt.Errorf("error encountered inputting commit to commits table: %v", err)
	}
	return nil
}

// GetAllRowsAscending returns the rows in the commits table sorted in ascending order
func GetAllRowsAscending() ([]Commit, error) {
	var commits []Commit
	rows, err := db.Query("SELECT id,repo_id,user_id,date,sha,notes FROM commits WHERE date IS NOT NULL ORDER BY date ASC")
	if err != nil {
		return commits, fmt.Errorf("error querying commits table for rows: %v", err)
	}
	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {
			log.Errorf("error closing sql rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var c commitWithNulls
		err := rows.Scan(&c.ID, &c.RepoID, &c.UserID, &c.Date, &c.SHA, &c.Notes)
		if err != nil {
			return commits, fmt.Errorf("error scanning row for commits table: %v", err)
		}

		nonNullCommit := convertSQLCommitToCommit(c)
		commits = append(commits, nonNullCommit)
	}
	if err := rows.Err(); err != nil {
		return commits, fmt.Errorf("error encountered iterating through commit rows: %v", err)
	}

	return commits, nil
}

// GetAllRowsByUserID returns the rows in the commits table that belong to a specific user ID
func GetAllRowsByUserID(uid int) ([]Commit, error) {
	var commits []Commit
	rows, err := db.Query("SELECT id,repo_id,user_id,date,sha,notes FROM commits WHERE user_id = ?", uid)
	if err != nil {
		return commits, fmt.Errorf("error querying commits table for rows: %v", err)
	}
	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {
			log.Errorf("error closing sql rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var c commitWithNulls
		err := rows.Scan(&c.ID, &c.RepoID, &c.UserID, &c.Date, &c.SHA, &c.Notes)
		if err != nil {
			return commits, fmt.Errorf("error scanning row for commits table: %v", err)
		}

		nonNullCommit := convertSQLCommitToCommit(c)
		commits = append(commits, nonNullCommit)
	}
	if err := rows.Err(); err != nil {
		return commits, fmt.Errorf("error encountered iterating through commit rows: %v", err)
	}

	return commits, nil
}

// DeleteRow deletes one row in the commits table by ID
// this will only be used to delete bot user activity once detected
func DeleteRow(id int) error {
	_, err := db.Exec(`DELETE FROM commits WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("error encountered deleting row for commits table: %v", err)
	}

	return nil
}

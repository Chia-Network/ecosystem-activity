package repos

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
	log "github.com/sirupsen/logrus"
)

// Repo represents all columns in one repo entry in the repos table
type Repo struct {
	ID              int
	Owner           string
	Repo            string
	ImportedThrough time.Time
	FirstCommit     time.Time
	LastCommit      time.Time
	Notes           string
}

// repoWithNulls is a helper struct for mysql rows that may contain null fields
// a null field using the mysql database driver won't scan into the appropriate field
type repoWithNulls struct {
	ID              sql.NullInt64
	Owner           sql.NullString
	Repo            sql.NullString
	ImportedThrough sql.NullTime
	FirstCommit     sql.NullTime
	LastCommit      sql.NullTime
	Notes           sql.NullString
}

// convertSQLRepoToRepo handles the internal conversion between an sql row response and a user-friendly repo struct
// because Go's sql package errors when scanning nil columns in a row
func convertSQLRepoToRepo(r repoWithNulls) Repo {
	var repo Repo
	if r.ID.Valid {
		repo.ID = int(r.ID.Int64)
	}
	if r.Owner.Valid {
		repo.Owner = r.Owner.String
	}
	if r.Repo.Valid {
		repo.Repo = r.Repo.String
	}
	if r.ImportedThrough.Valid {
		repo.ImportedThrough = r.ImportedThrough.Time
	}
	if r.FirstCommit.Valid {
		repo.FirstCommit = r.FirstCommit.Time
	}
	if r.LastCommit.Valid {
		repo.LastCommit = r.LastCommit.Time
	}
	if r.Notes.Valid {
		repo.Notes = r.Notes.String
	}
	return repo
}

// GetRowsByOwnerAndRepo returns the rows where the owner and repo both match (should be one row)
func GetRowsByOwnerAndRepo(owner, repo string) ([]Repo, error) {
	var repos []Repo
	rows, err := db.Query("SELECT id,owner,repo,imported_through,first_commit,last_commit,notes FROM repos WHERE owner = ? AND repo = ?", owner, repo)
	if err != nil {
		return repos, fmt.Errorf("error querying repos table for rows by owner \"%s\" and repo \"%s\": %v", owner, repo, err)
	}
	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {
			log.Errorf("error closing sql rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var r repoWithNulls
		err := rows.Scan(&r.ID, &r.Owner, &r.Repo, &r.ImportedThrough, &r.FirstCommit, &r.LastCommit, &r.Notes)
		if err != nil {
			return repos, fmt.Errorf("error scanning row for owner \"%s\" and repo \"%s\": %v", owner, repo, err)
		}

		nonNullRepo := convertSQLRepoToRepo(r)
		repos = append(repos, nonNullRepo)
	}
	if err := rows.Err(); err != nil {
		return repos, fmt.Errorf("error encountered iterating through rows for owner \"%s\" and repo \"%s\": %v", owner, repo, err)
	}

	return repos, nil
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(repo Repo) error {
	_, err := db.Exec(`INSERT INTO repos (owner,repo) VALUES(?, ?);`, repo.Owner, repo.Repo)
	if err != nil {
		return fmt.Errorf("error adding repo to repos table for \"%s\" and repo \"%s\": %v", repo.Owner, repo.Repo, err)
	}

	return nil
}

// UpdateLastCommitByID accepts a row ID and time object and updates the matching row's last_commit column to the timestamp
func UpdateLastCommitByID(id int, ts time.Time) error {
	_, err := db.Exec(`UPDATE repos SET last_commit=? WHERE id=?;`, ts.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("error encountered updating last_commit on row ID %d: %v", id, err)
	}
	return err
}

// UpdateFirstCommitByID accepts a row ID and time object and updates the matching row's first_commit column to the timestamp
func UpdateFirstCommitByID(id int, ts time.Time) error {
	_, err := db.Exec(`UPDATE repos SET first_commit=? WHERE id=?;`, ts.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("error encountered updating first_commit on row ID %d: %v", id, err)
	}
	return err
}

// UpdateImportedThroughByID accepts a row ID and time object and updates the matching row's imported_through column to the timestamp
func UpdateImportedThroughByID(id int, ts time.Time) error {
	_, err := db.Exec(`UPDATE repos SET imported_through=? WHERE id=?;`, ts.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("error encountered updating imported_through on row ID %d: %v", id, err)
	}
	return err
}

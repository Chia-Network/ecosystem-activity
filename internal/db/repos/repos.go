package repos

import (
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
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

// SetNewRecord inserts one new record into the table
func SetNewRecord(r Repo) ([]Repo, error) {
	var repos []Repo
	rows, err := db.Query(`INSERT INTO repos (owner,repo,notes) VALUES(?, ?, ?, ?, ?);`, r.Owner, r.Repo, r.Notes)
	if err != nil {
		return []Repo{}, fmt.Errorf("error adding repo to repos table for \"%s\" and repo \"%s\": %v", r.Owner, r.Repo, err)
	}
	defer rows.Close()

	for rows.Next() {
		var repo Repo
		err := rows.Scan(&repo.ID, &repo.Owner, &repo.Repo, &repo.ImportedThrough, &repo.FirstCommit, &repo.LastCommit, &repo.Notes)
		if err != nil {
			return []Repo{}, fmt.Errorf("error scanning row for owner \"%s\" and repo \"%s\": %v", r.Owner, r.Repo, err)
		}
		repos = append(repos, repo)
	}
	if err := rows.Err(); err != nil {
		return []Repo{}, fmt.Errorf("error encountered iterating through rows for owner \"%s\" and repo \"%s\": %v", r.Owner, r.Repo, err)
	}

	return repos, nil
}

// GetRowsByOwnerAndRepo returns the rows where the owner and repo both match (should be one row)
func GetRowsByOwnerAndRepo(owner, repo string) ([]Repo, error) {
	var repos []Repo
	rows, err := db.Query("SELECT id,owner,repo,imported_through,first_commit,last_commit,notes FROM repos WHERE owner = ? AND repo = ?", owner, repo)
	if err != nil {
		return repos, fmt.Errorf("error querying repos table for rows by owner \"%s\" and repo \"%s\": %v", owner, repo, err)
	}
	defer rows.Close()

	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.ID, &r.Owner, &r.Repo, &r.ImportedThrough, &r.FirstCommit, &r.LastCommit, &r.Notes)
		if err != nil {
			return repos, fmt.Errorf("error scanning row for owner \"%s\" and repo \"%s\": %v", owner, repo, err)
		}
		repos = append(repos, r)
	}
	if err := rows.Err(); err != nil {
		return repos, fmt.Errorf("error encountered iterating through rows for owner \"%s\" and repo \"%s\": %v", owner, repo, err)
	}

	return repos, nil
}

// UpdateLastCommitByID accepts a row ID and time object and updates the matching row's last_commit column to the timestamp
func UpdateLastCommitByID(id int, ts time.Time) error {
	_, err := db.Exec(`UPDATE repos SET last_commit='?' WHERE id='?';`, ts.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("error encountered updating last_commit on row ID %d: %v", id, err)
	}
	return err
}

// UpdateFirstCommitByID accepts a row ID and time object and updates the matching row's first_commit column to the timestamp
func UpdateFirstCommitByID(id int, ts time.Time) error {
	_, err := db.Exec(`UPDATE repos SET first_commit='?' WHERE id='?';`, ts.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("error encountered updating first_commit on row ID %d: %v", id, err)
	}
	return err
}

// UpdateImportedThroughByID accepts a row ID and time object and updates the matching row's imported_through column to the timestamp
func UpdateImportedThroughByID(id int, ts time.Time) error {
	_, err := db.Exec(`UPDATE repos SET imported_through='?' WHERE id='?';`, ts.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("error encountered updating imported_through on row ID %d: %v", id, err)
	}
	return err
}

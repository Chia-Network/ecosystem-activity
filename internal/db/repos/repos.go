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
	CreatedAt       time.Time
	ImportedThrough time.Time
	FirstCommit     time.Time
	LastCommit      time.Time
	Notes           string
}

// GetRowsByOwnerAndRepo returns the rows where the owner and repo both match (should be one row)
func GetRowsByOwnerAndRepo(owner, repo string) ([]Repo, error) {
	var repos []Repo
	rows, err := db.Query("SELECT id,owner,repo,created_at,imported_through,first_commit,last_commit,notes FROM repos WHERE owner = ? AND repo = ?", owner, repo)
	if err != nil {
		return repos, fmt.Errorf("error querying repos table for rows by owner \"%s\" and repo \"%s\": %v", owner, repo, err)
	}
	defer rows.Close()

	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.ID, &r.Owner, &r.Repo, &r.CreatedAt, &r.ImportedThrough, &r.FirstCommit, &r.LastCommit, &r.Notes)
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

package sortedcommits

import (
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
)

// SortedCommit represents all columns in one sorted_commit entry in the sorted_commits table
type SortedCommit struct {
	ID       int
	CommitID int
	Date     time.Time
}

// ResetAllRecords this function would be called to reset the sorted_commits table for reorganization
// First deletes all rows in the table and then resets the auto_increment counter for the id column
func ResetAllRecords() error {
	_, err := db.Exec(`DELETE FROM sorted_commits;`)
	if err != nil {
		return fmt.Errorf("error encountered deleting records for sorted_commits table: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE sorted_commits AUTO_INCREMENT = 1;`)
	if err != nil {
		return fmt.Errorf("error encountered reseting auto_increment for sorted_commits table: %v", err)
	}
	return nil
}

// SetNewRecord inserts one new record into the table
func SetNewRecord(c SortedCommit) error {
	if c.CommitID == 0 || c.Date.IsZero() {
		// Safe to just return here, we sanity checked the input and it was bad but we don't need to gate anything with this
		return nil
	}
	_, err := db.Exec(`INSERT INTO sorted_commits (commit_id,date) VALUES(?, ?);`, c.CommitID, c.Date.Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("error encountered inputting commit to commits table: %v", err)
	}
	return nil
}

package sorter

import (
	"github.com/chia-network/ecosystem-activity/internal/db/commits"
	sortedcommits "github.com/chia-network/ecosystem-activity/internal/db/sorted_commits"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

// Schedule creates a cron for refreshing the sorted commit table
func Schedule(schedule string) {
	log.Infof("registering sorter cron with schedule \"%s\"", schedule)
	c := cron.New()
	err := c.AddFunc(schedule, RunSortedCommits)
	if err != nil {
		log.Errorf("error encountered registering sorter cron: %v", err)
	}
	c.Start()
}

// RunSortedCommits deletes all records in the sorted_commits table, restarts the auto incrementer, and adds all the commits in ascending order from the commits table
func RunSortedCommits() {
	log.Info("Running the commit sorter for the sorted_commits table")

	// Gather all commits in the commits table in ascending order
	allCommitsAsc, err := commits.GetAllRowsAscending()
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Found %d commits to add to the sorted_commits table", len(allCommitsAsc))

	// Deletes all records in the sorted commits table
	// This is obviously an operation that can't be reversed except with an import,
	// the loop below should re-add the rows in the correct order
	err = sortedcommits.ResetAllRecords()
	if err != nil {
		log.Error(err)
		return
	}

	// Create each record in the ascending datetime order as returned by the commits table
	for i, commit := range allCommitsAsc {
		log.Debugf("Adding commit ID %d, iteration %d", commit.ID, i)
		err = sortedcommits.SetNewRecord(sortedcommits.SortedCommit{
			CommitID: commit.ID,
			Date:     commit.Date,
		})
		if err != nil {
			log.Error(err)
			return
		}
	}
}

package collector

import (
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db/commits"
	"github.com/chia-network/ecosystem-activity/internal/db/repos"
	"github.com/chia-network/ecosystem-activity/internal/db/users"
	gh "github.com/chia-network/ecosystem-activity/internal/github"

	"github.com/google/go-github/v52/github"
	log "github.com/sirupsen/logrus"
)

// A github repo was identified, will query commit data using the github API
func githubRepo(owner string, repo string) {
	ownerRepoString := fmt.Sprintf("%s/%s", owner, repo)

	// Get the row data for this repo in the repos table (makes a new row if one does not exist)
	repoRow, repoInTable, err := getRepoRow(owner, repo)
	if err != nil {
		log.Error(err)
		return
	}

	// Get search start time by checking if the repo was already searched and using the last search time/datestamp if it was, use Chia Network incorporation date as genesis if not
	var searchStart time.Time
	if !repoInTable {
		searchStart = time.Date(2017, time.August, 1, 0, 0, 0, 0, time.UTC)
	} else {
		searchStart = repoRow.ImportedThrough
	}

	// Search end time is always just now in UTC, but saving the timestamp here to ensure accurate timestamps in the `repos` table's `imported_through` column
	searchEnd := time.Now().UTC()

	// Query repository commits between a start and end date
	cmts, statusCode, err := gh.ListRepositoryCommits(owner, repo, searchStart, searchEnd)
	if statusCode == 404 {
		log.Warnf("Repo %s returned a 404", ownerRepoString)
		return
	}
	if err != nil {
		log.Errorf("Failed to get commit list for %s/%s with error: %v", owner, repo, err)
		return
	}

	// Set repo in table because it did not 404
	if !repoInTable {
		repoRow, err = setRepoRow(repos.Repo{
			Owner: owner,
			Repo:  repo,
		})
		if err != nil {
			log.Error(err)
			return
		}
	}

	log.Debugf("Successfully queried commits for repo %s/%s, found %d commits", owner, repo, len(cmts))

	// For each commit we need to identify important data from the API response and submit it to the db
	var latestCommit, earliestCommit time.Time
	for _, commit := range cmts {
		commitSHA, err := getCommitSHA(commit)
		if err != nil {
			log.Errorf("failed to read commit sha data for %s: %v", ownerRepoString, err)
			continue
		}
		if commitSHA == "" {
			log.Errorf("commit data was not nil but no SHA returned from API for repo %s: %v", ownerRepoString, err)
			continue
		}

		// Since this tool doesn't filter out bot users itself, bot users will need to be filtered out from queries in Grafana
		// Unfortunately filtering bot users from results in this tool would probably be prone to false positives and/or negatives
		commitAuthorLogin, err := getCommitAuthorLogin(commit)
		if err != nil {
			log.Errorf("failed to read commit author login for %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}
		if commitAuthorLogin == "" {
			log.Errorf("commit data was not nil but no author login returned from API for repo %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}

		commitTimestamp, err := getCommitDate(commit)
		if err != nil {
			log.Errorf("failed to read commit date for %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}
		if commitTimestamp.IsZero() {
			log.Errorf("commit data was not nil but no commit timestamp returned from API for repo %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}

		// Check if commit author already exists in users table
		userRow, ok, err := getUserRow(commitAuthorLogin)
		if err != nil {
			log.Error(err)
			continue
		}
		// If user did exist, check if commit timestamp is later than `last_commit` or earlier than `first_commit`, if so, update the row
		if ok {
			if userRow.FirstCommit.After(commitTimestamp) || userRow.FirstCommit.IsZero() {
				err = users.UpdateFirstCommitByUsername(commitAuthorLogin, commitTimestamp)
				if err != nil {
					log.Error(err)
				}
			}
			if userRow.LastCommit.Before(commitTimestamp) || userRow.LastCommit.IsZero() {
				err = users.UpdateLastCommitByUsername(commitAuthorLogin, commitTimestamp)
				if err != nil {
					log.Error(err)
				}
			}
		}
		// If user did not exist, add them to users table, this commit can be the first and last
		if !ok {
			userRow, err = setUserRow(users.User{
				Username:    commitAuthorLogin,
				FirstCommit: commitTimestamp,
				LastCommit:  commitTimestamp,
			})
			if err != nil {
				log.Error(err)
				continue
			}
		}

		// Add commit to commits table
		err = commits.SetNewRecord(commits.Commit{
			RepoID: repoRow.ID,
			UserID: userRow.ID,
			Date:   commitTimestamp,
			SHA:    commitSHA,
		})
		if err != nil {
			log.Errorf("error encountered submitting commit record to commits table: %v", err)
			continue
		}

		// Check if earliest commit or latest commit from this batch of commits
		if earliestCommit.IsZero() || earliestCommit.After(commitTimestamp) {
			earliestCommit = commitTimestamp
		}
		if latestCommit.IsZero() || latestCommit.Before(commitTimestamp) {
			latestCommit = commitTimestamp
		}
	}

	// Update repos row. If earliest commit is earlier than `first_commit` or `first_commit` is empty, set to this commit's timestamp.
	// If latest commit is later than `last_commit` or `last_commit` is empty, set to this commit's timestamp
	if !earliestCommit.IsZero() {
		if repoRow.FirstCommit.After(earliestCommit) || repoRow.FirstCommit.IsZero() {
			log.Debugf("setting first commit for repo %s/%s. current first commit %v. new first commit %v.", repoRow.Owner, repoRow.Repo, repoRow.FirstCommit, earliestCommit)
			err = repos.UpdateFirstCommitByID(repoRow.ID, earliestCommit)
			if err != nil {
				log.Error(err)
			}
		}
	}
	if !latestCommit.IsZero() {
		if latestCommit.After(repoRow.LastCommit) || repoRow.LastCommit.IsZero() {
			log.Debugf("setting last commit for repo %s/%s. current last commit %v. new last commit %v.", repoRow.Owner, repoRow.Repo, repoRow.LastCommit, latestCommit)
			err = repos.UpdateLastCommitByID(repoRow.ID, latestCommit)
			if err != nil {
				log.Error(err)
			}
		}
	}

	// Update repo's `imported_through` column as we finished importing these commits
	err = repos.UpdateImportedThroughByID(repoRow.ID, searchEnd)
	if err != nil {
		log.Error(err)
		return
	}
}

// getUserRow looks up a user by username in the users table.
// Returns the user row object, a boolean value to signal if a single user was found in the table, and an optional error
func getUserRow(u string) (users.User, bool, error) {
	var userRow users.User
	rows, err := users.GetRowsByUsername(u)
	if err != nil {
		return userRow, false, err
	}

	if len(rows) > 1 {
		return userRow, false, fmt.Errorf("multiple rows found for user %s -- this would signify an unexpected condition, please check users table", u)
	}

	if len(rows) == 0 {
		return userRow, false, nil
	}

	return rows[0], true, nil
}

func setUserRow(u users.User) (users.User, error) {
	err := users.SetNewRecord(u)
	if err != nil {
		return users.User{}, err
	}

	userRow, _, err := getUserRow(u.Username)
	if err != nil {
		return users.User{}, err
	}

	return userRow, nil
}

func getRepoRow(owner string, repo string) (repos.Repo, bool, error) {
	var repoRow repos.Repo
	rows, err := repos.GetRowsByOwnerAndRepo(owner, repo)
	if err != nil {
		return repoRow, false, err
	}

	if len(rows) > 1 {
		return repoRow, false, fmt.Errorf("multiple rows found for %s/%s -- this would signify an unexpected condition, please check repos table", owner, repo)
	}

	if len(rows) == 0 {
		return repoRow, false, nil
	}

	return rows[0], true, nil
}

func setRepoRow(r repos.Repo) (repos.Repo, error) {
	err := repos.SetNewRecord(r)
	if err != nil {
		return repos.Repo{}, err
	}

	repoRow, _, err := getRepoRow(r.Owner, r.Repo)
	if err != nil {
		return repos.Repo{}, err
	}

	return repoRow, nil
}

func getCommitSHA(commit *github.RepositoryCommit) (string, error) {
	if commit.SHA != nil {
		return *commit.SHA, nil
	}
	if commit.Commit.SHA != nil {
		return *commit.Commit.SHA, nil
	}
	return "", fmt.Errorf("both RepositoryCommit.SHA and Commit.SHA were nil")
}

func getCommitAuthorLogin(commit *github.RepositoryCommit) (string, error) {
	// Check to make sure commit author's login is not nil
	if commit.Author != nil {
		if commit.Author.Login != nil {
			return *commit.Author.Login, nil
		}
	}

	if commit.Commit != nil {
		if commit.Commit.Author != nil {
			if commit.Commit.Author.Login != nil {
				return *commit.Commit.Author.Login, nil
			}
		}
	}

	return "", fmt.Errorf("both RepositoryCommit.Author.Login and Commit.Author.Login were nil during commit author login check")
}

func getCommitDate(commit *github.RepositoryCommit) (time.Time, error) {
	if commit.Commit != nil {
		if commit.Commit.Author != nil {
			return commit.Commit.Author.GetDate().Time, nil
		}
	}
	return time.Time{}, fmt.Errorf("both RepositoryCommit.Author and Commit.Author were nil during commit timestamp check")
}

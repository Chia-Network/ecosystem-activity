package collector

import (
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db/repos"
	gh "github.com/chia-network/ecosystem-activity/internal/github"

	"github.com/google/go-github/v52/github"
	log "github.com/sirupsen/logrus"
)

// A github repo was identified, will query commit data using the github API
func githubRepo(owner string, repo string) {
	ownerRepoString := fmt.Sprintf("%s/%s", owner, repo)

	/*
		TODO
		Order of Operations:
		1) Check if repo already exists in repos table.
		2a) If it does, get the row ID and save for later.
		2b) If repo didn't exist in repos table, query github for repo data, specifically need a repo creation date. (IF repo 404s, log it and return.)
			Then create new row in repos table leaving `imported_through`, `first_commit`, and `last_commit` blank for now.
	*/

	// Get search start time by checking if the repo was already searched and using the last search time/datestamp if it was, use Chia Network incorporation date as genesis if not
	searchStart, err := getSearchStartTime(owner, repo)
	if err != nil {
		log.Error(err)
		return
	}
	// Search end time is always just now in UTC, but saving the timestamp here to ensure accurate timestamps in the `repos` table's `imported_through` column
	searchEnd := time.Now().UTC()

	// Query repository commits between a start and end date
	commits, statusCode, err := gh.ListRepositoryCommits(owner, repo, searchStart, searchEnd)
	if statusCode == 404 {
		log.Warnf("Repo %s returned a 404", ownerRepoString)
	}
	if err != nil {
		log.Errorf("Failed to get commit list for %s/%s with error: %v", owner, repo, err)
		return
	}

	log.Debugf("Successfully queried commits for repo %s/%s, found %d commits", owner, repo, len(commits))

	// For each commit we need to identify important data from the API response and submit it to the db
	for _, commit := range commits {
		commitSHA, err := getCommitSHA(commit)
		if err != nil {
			log.Errorf("failed to read commit sha data for %s: %v", ownerRepoString, err)
			continue
		}

		// Since this tool doesn't filter out bot users itself, bot users will need to be filtered out from queries in Grafana
		// Unfortunately filtering bot users from results in this tool would probably be prone to false positives and/or negatives
		commitAuthorLogin, err := getCommitAuthorLogin(commit)
		if err != nil {
			log.Errorf("failed to read commit author login for %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}

		commitAuthorEmail, err := getCommitAuthorEmail(commit)
		if err != nil {
			log.Errorf("failed to read commit author email for %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}

		commitAuthorName, err := getCommitAuthorName(commit)
		if err != nil {
			log.Errorf("failed to read commit author name for %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}

		commitTimestamp, err := getCommitDate(commit)
		if err != nil {
			log.Errorf("failed to read commit date for %s, sha %s: %v", ownerRepoString, commitSHA, err)
			continue
		}

		/*
			TODO
			Order of operations:
			1) Check if user already exists in users table. If not query github for user data, specifically need a user creation date.
			2a) IF user does exist in users table, check if commit timestamp is newer than `last_commit` field for user. If so, update it to this commit's timestamp. Save row ID for later
			2b) IF user did not exist in users table add them and save row ID for later
			3) Insert new commit row to commits table
			4) Update repos row. `imported_through` set to searchEnd var. IF this commit is earlier than `first_commit` or `first_commit` is empty, set to this commit's timestamp.
				IF this commit is later than `last_commit` or `last_commit` is empty, set to this commit's timestamp
		*/
	}
}

func getSearchStartTime(owner string, repo string) (time.Time, error) {
	// Check if repo already exists in repos table
	r, err := repos.GetRowsByOwnerAndRepo(owner, repo)
	if err != nil {
		return time.Time{}, err
	}

	// If multiple rows found, something weird happened. Log an error and leave it.
	if len(r) > 1 {
		return time.Time{}, fmt.Errorf("multiple rows found for %s/%s -- this would signify an unexpected condition, please check repos table", owner, repo)
	}

	// If one row was found, it was in the repos table, so this interval should start where the last interval ended
	if len(r) == 1 {
		return r[0].ImportedThrough, nil
	}

	// If no rows found, using the Chia incorporated date as a genesis
	// Source: https://www.chia.net/faq/ "Chia was incorporated in August of 2017..."
	return time.Date(2017, time.August, 1, 0, 0, 0, 0, time.UTC), nil
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

func getCommitAuthorEmail(commit *github.RepositoryCommit) (string, error) {
	if commit.Commit != nil {
		if commit.Commit.Author != nil {
			return commit.Commit.Author.GetEmail(), nil
		}
	}

	return "", fmt.Errorf("Commit.Author was nil during commit author email check")
}

func getCommitAuthorName(commit *github.RepositoryCommit) (string, error) {
	if commit.Commit != nil {
		if commit.Commit.Author != nil {
			return commit.Commit.Author.GetName(), nil
		}
	}

	return "", fmt.Errorf("Commit.Author was nil during commit author name check")
}

func getCommitDate(commit *github.RepositoryCommit) (string, error) {
	timeFormat := "2006-01-02 15:04:05"
	// Check to make sure commit is not nil
	if commit.Commit != nil {
		if commit.Commit.Author != nil {
			return commit.Commit.Author.GetDate().Time.Format(timeFormat), nil
		}
	}
	return "", fmt.Errorf("both RepositoryCommit.Author and Commit.Author were nil during commit timestamp check")
}

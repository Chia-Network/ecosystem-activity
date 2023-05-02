package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/ecosystem-activity/internal/db"
	gh "github.com/chia-network/ecosystem-activity/internal/github"
)

const mysqlDateFormat = "2006-01-02 15:04:05"

// backfillCmd represents the backfill command
var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Fills missing data for repos and users",
	Run: func(cmd *cobra.Command, args []string) {
		// Init github package with auth token
		gh.Init(viper.GetString("github-token"))

		// Init db package
		err := db.Init(viper.GetString("mysql-host"), viper.GetString("mysql-database"), viper.GetString("mysql-user"), viper.GetString("mysql-password"))
		if err != nil {
			log.Error(err)
		}

		rows, err := db.Query("select id, owner, repo from repos where created_at IS NULL or first_commit IS NULL or last_commit IS NULL")
		if err != nil {
			log.Fatalf("Error querying repos: %s\n", err.Error())
		}
		for rows.Next() {
			var (
				id    int
				owner string
				repo  string
			)
			if err := rows.Scan(&id, &owner, &repo); err != nil {
				log.Fatalf("Error scanning row: %s\n", err.Error())
			}

			log.Printf("Looking up data on repo %s/%s\n", owner, repo)
			repository, _, err := gh.GetRepository(owner, repo)
			if err != nil {
				log.Printf("Error getting repository: %s\n", err.Error())
				continue
			}

			createdAt := repository.CreatedAt.Format(mysqlDateFormat)
			firstCommit := getFirstCommit(owner, repo)
			lastCommit := getLastCommit(owner, repo)

			if createdAt == "" || firstCommit == "" || lastCommit == "" {
				continue
			}
			_, err = db.Exec("Update repos set created_at = ?, first_commit = ?, last_commit = ? where id = ?", createdAt, firstCommit, lastCommit, id)
			if err != nil {
				log.Fatalf("Error updating record for %s/%s: %s\n", owner, repo, err.Error())
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(backfillCmd)
}

// Gets the first commit based on commit data in this DB
func getFirstCommit(owner, repo string) string {
	repoID, err := getRepoID(owner, repo)
	if err != nil {
		return ""
	}

	rows, err := db.Query("select date from commits where repo_id = ? order by date asc limit 1", repoID)
	if err != nil {
		return ""
	}
	rows.Next()
	var date string
	err = rows.Scan(&date)
	if err != nil {
		return ""
	}
	return date
}

// Gets the last commit based on commit data in this DB
func getLastCommit(owner, repo string) string {
	repoID, err := getRepoID(owner, repo)
	if err != nil {
		return ""
	}

	rows, err := db.Query("select date from commits where repo_id = ? order by date desc limit 1", repoID)
	if err != nil {
		return ""
	}
	rows.Next()
	var date string
	err = rows.Scan(&date)
	if err != nil {
		return ""
	}
	return date
}

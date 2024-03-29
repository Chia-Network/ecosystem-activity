package cmd

import (
	"database/sql"
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/ecosystem-activity/internal/db"
	"github.com/chia-network/ecosystem-activity/internal/utils"
)

var (
	file string
)

// importCommitsCmd represents the importCommits command
var importCommitsCmd = &cobra.Command{
	Use:   "import-commits",
	Short: "Imports commits from a CSV generated by the old reporting tool",
	Long: `Imports commits from a CSV generated by the old reporting tool.

This importer expects that the repos and users tables already have data, and the importer will look up repo and user 
IDs from that table based on the Owner and Repository fields in the CSV.

CSV is expected to have the following fields:
Owner,Repository,Commit Author, Commit SHA, Commit Date`,
	Run: func(cmd *cobra.Command, args []string) {
		// Init db package
		err := db.Init(viper.GetString("mysql-host"), viper.GetString("mysql-database"), viper.GetString("mysql-user"), viper.GetString("mysql-password"))
		if err != nil {
			log.Error(err)
		}

		log.Printf("Importing %s\n", file)
		f, err := os.Open(file)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Fatalf("Error closing file: %s\n", err.Error())
			}
		}(f)

		csvReader := csv.NewReader(f)
		data, err := csvReader.ReadAll()
		if err != nil {
			log.Fatalf("Error reading CSV: %s\n", err.Error())
		}

		for rowNum, line := range data {
			if rowNum == 0 {
				continue
			}
			owner := line[0]
			repo := line[1]
			commitAuthor := line[2]
			commitSHA := line[3]
			commitDate := line[4]

			if isPossibleBot := utils.MatchesBot(commitAuthor); isPossibleBot {
				// Skip this commit since it matched a bot account username matcher
				continue
			}

			repoID, err := getRepoID(owner, repo)
			if err != nil {
				log.Fatalf("Error getting repo %s/%s: %s\n", owner, repo, err.Error())
			}

			authorID, err := getCommitAuthorID(commitAuthor)
			if err != nil {
				log.Fatalf("Error getting commit author %s: %s\n", commitAuthor, err.Error())
			}

			log.Printf("%s/%s (%d) %s (%d) %s %s\n", owner, repo, repoID, commitAuthor, authorID, commitSHA, commitDate)

			_, err = db.Exec("INSERT INTO commits (repo_id, user_id, date, sha) VALUES (?, ?, ?, ?)", repoID, authorID, commitDate, commitSHA)
			if err != nil {
				log.Fatalf("Error writing to DB: %s\n", err.Error())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(importCommitsCmd)

	importCommitsCmd.Flags().StringVar(&file, "file", "", "The file to import")
}

func getCommitAuthorID(username string) (int, error) {
	var id int
	result, err := db.Query("select id from users where username = ?", username)
	if err != nil {
		return 0, err
	}
	defer func(*sql.Rows) {
		err := result.Close()
		if err != nil {
			log.Errorf("error closing sql rows: %v", err)
		}
	}(result)

	result.Next()
	err = result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func getRepoID(owner, repo string) (int, error) {
	var id int
	result, err := db.Query("select id from repos where owner = ? and repo = ?", owner, repo)
	if err != nil {
		return 0, err
	}
	defer func(*sql.Rows) {
		err := result.Close()
		if err != nil {
			log.Errorf("error closing sql rows: %v", err)
		}
	}(result)

	result.Next()
	err = result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

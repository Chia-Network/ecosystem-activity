package cmd

import (
	"fmt"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/db"
	"github.com/chia-network/ecosystem-activity/internal/db/commits"
	sortedcommits "github.com/chia-network/ecosystem-activity/internal/db/sorted_commits"
	"github.com/chia-network/ecosystem-activity/internal/db/users"
	"github.com/chia-network/ecosystem-activity/internal/sorter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startDate     string    // The unformatted string given from flag
	startDateTime time.Time // startDate parsed in to a time object
)

// deleteBotRowsCmd represents the deleteBotRows command
var deleteBotRowsCmd = &cobra.Command{
	Use:   "delete-bots",
	Short: "Runs the bot deleting script",
	Long: `New bots may be found that don't get filtered by the bot finding measures currently set.

This uses the bot matching utility functions to discover bots in the prod dataset, and deletes bot related data as it's irrelevant to developer metrics.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Format start date string to time
		var err error
		startDateTime, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			log.Fatalf("parsed time from flag returned an error: %v", err)
		}

		// Init db package
		err = db.Init(viper.GetString("mysql-host"), viper.GetString("mysql-database"), viper.GetString("mysql-user"), viper.GetString("mysql-password"))
		if err != nil {
			log.Fatal(err)
		}

		// Get all user rows that match listed bot matchers
		botUsers, err := users.GetBotUserRows()
		if err != nil {
			log.Fatal(err)
		}

		// Log users to be deleted and wait for implicit confirmation
		logUsers(botUsers)
		time.Sleep(30 * time.Second)

		for _, u := range botUsers {
			// Get all commit rows for the user
			cmts, err := commits.GetAllRowsByUserID(u.ID)
			if err != nil {
				log.Error(err)
				continue
			}

			// Loop through each commit
			deletedAll := true // This will be switched to false if a commit is ever skipped
			for _, cmt := range cmts {
				// Check if date is after start date, skip if not
				if cmt.Date.Before(startDateTime) {
					log.Infof("skipping commit %s for user %s, commit date %s was before start time %s", cmt.SHA, u.Username, cmt.Date.Format("2006-01-02 15:04:05"), startDateTime.Format("2006-01-02 15:04:05"))
					deletedAll = false
					continue
				}
				log.Infof("deleting commit %s for user %s, commit date %s was after start time %s", cmt.SHA, u.Username, cmt.Date.Format("2006-01-02 15:04:05"), startDateTime.Format("2006-01-02 15:04:05"))

				// Delete the row in sorted_commits table
				err = sortedcommits.DeleteRow(cmt.ID)
				if err != nil {
					log.Error(err)
					deletedAll = false
					continue
				}

				// Delete the row in commits table
				err = commits.DeleteRow(cmt.ID)
				if err != nil {
					log.Error(err)
					deletedAll = false
					continue
				}
			}

			// Delete the user row if all rows for the user in the commits tables were deleted
			if deletedAll {
				log.Infof("deleting user %s because all of their commits were deleted", u.Username)
				err = users.DeleteRow(u.ID)
				if err != nil {
					log.Error(err)
				}
			}
		}

		// Run the adhoc sorted commits function, this will usually take a while in prod, grab a cup of coffee
		sorter.RunSortedCommits()
	},
}

func init() {
	rootCmd.AddCommand(deleteBotRowsCmd)
	deleteBotRowsCmd.Flags().StringVar(&startDate, "start-date", "", "The date to start looking for bot data in YYYY-MM-DD format")
}

func logUsers(users []users.User) {
	fmt.Println("Deleting user and commit activity for the following users:")
	for _, u := range users {
		fmt.Printf(" * %s\n", u.Username)
	}
	fmt.Println("Waiting 30 seconds before continuing...")
}

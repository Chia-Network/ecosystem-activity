package cmd

import (
	"github.com/chia-network/ecosystem-activity/internal/db"
	"github.com/chia-network/ecosystem-activity/internal/sorter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// adhocSortedCommitsCmd represents the adhocSortedCommits command
var adhocSortedCommitsCmd = &cobra.Command{
	Use:   "import-commits",
	Short: "Runs the sorted commits function ad-hoc",
	Long: `Run an ad-hoc iteration of the sorted commits function.

This deletes the rows in the sorted_commits table, gets a list of commits from the commits table, and adds them back to the sorted_commits table in ascending order.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Init db package
		err := db.Init(viper.GetString("mysql-host"), viper.GetString("mysql-database"), viper.GetString("mysql-user"), viper.GetString("mysql-password"))
		if err != nil {
			log.Error(err)
		}

		// Run ad-hoc
		sorter.RunSortedCommits()
	},
}

func init() {
	rootCmd.AddCommand(adhocSortedCommitsCmd)
}

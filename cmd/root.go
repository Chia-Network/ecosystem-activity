package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/chia-network/ecosystem-activity/internal/collector"
	"github.com/chia-network/ecosystem-activity/internal/config"
	"github.com/chia-network/ecosystem-activity/internal/db"
	gh "github.com/chia-network/ecosystem-activity/internal/github"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecosystem-activity",
	Short: "View stats on user commit activity for a set of repos over the lifespan of those repositories.",
	Run: func(cmd *cobra.Command, args []string) {
		// Init github package with auth token
		gh.Init(viper.GetString("github-token"))

		// Init db package
		err := db.Init(viper.GetString("mysql-host"), viper.GetString("mysql-database"), viper.GetString("mysql-user"), viper.GetString("mysql-password"))
		if err != nil {
			log.Error(err)
		}

		// Run collector, the main logic loop for this data collector tool
		go collector.Run(cfg, viper.GetInt("interval"))

		// Healthcheck handler
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			// TODO -- perhaps write a check here for the last time an import was done, and if earlier than such and such time, return 503 service unavailable
			// Currently this check only works if the whole application crashes, then the healthcheck requests will return an error, hopefully causing the container to respawn
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, "ok")
			if err != nil {
				log.Errorf("error writing to io writer: %v", err)
				return
			}
		})

		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Errorf("error returned from http ListenAndServe: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		cobra.CheckErr(rootCmd.Execute())
	}
}

func init() {
	var cfgFile string
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file (default: ./config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "How verbose the logs should be. panic, fatal, error, warn, info, debug, trace (default: info)")
	rootCmd.PersistentFlags().String("github-token", "", "A GitHub API token")
	rootCmd.PersistentFlags().Int("interval", 60, "An integer interval duration, specified in minutes, between collector runs")
	rootCmd.PersistentFlags().String("mysql-host", "", "The hostname to connect to for the mysql db")
	rootCmd.PersistentFlags().String("mysql-database", "", "The mysql database to use")
	rootCmd.PersistentFlags().String("mysql-user", "", "A mysql username to authenticate as, requires a password, see the `--mysql-password` flag")
	rootCmd.PersistentFlags().String("mysql-password", "", "A password for the corresponding mysql username, see the `--mysql-user` flag")
	cobra.OnInitialize(func() { initConfig(cfgFile) })

	err := viper.BindPFlag("github-token", rootCmd.PersistentFlags().Lookup("github-token"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("interval", rootCmd.PersistentFlags().Lookup("interval"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-host", rootCmd.PersistentFlags().Lookup("mysql-host"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-database", rootCmd.PersistentFlags().Lookup("mysql-database"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-user", rootCmd.PersistentFlags().Lookup("mysql-user"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = viper.BindPFlag("mysql-password", rootCmd.PersistentFlags().Lookup("mysql-password"))
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("ECOSYSTEM_ACTIVITY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// Unmarshal config to struct
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("unmarshalling config file: %v", err)
	}

	// Set log level for logrus
	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatalf("Error parsing log level: %s\n", err.Error())
	}
	log.Infof("Setting log level to: %s", viper.GetString("log-level"))
	log.SetLevel(level)
}

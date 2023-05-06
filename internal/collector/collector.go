package collector

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chia-network/ecosystem-activity/internal/config"
	gh "github.com/chia-network/ecosystem-activity/internal/github"

	log "github.com/sirupsen/logrus"
)

// List of repos subject to commit activity reporting
var repoList map[string]bool

// Run is the main logic loop for the collector service and accepts a config object and a resting interval duration in minutes for the collector service loop
func Run(cfg config.Config, interval int) {
	// Assemble full repo list from config, querying git remote site's specified orgs for additional repositories
	repoList = make(map[string]bool)
	err := createRepoList(cfg)
	if err != nil {
		log.Fatalf("couldn't put together a repo list from config: %v", err)
	}

	// This loop should continue indefinitely, being ran in its own goroutine, called from the cmd package
	for {
		// Loop through all repos in map, retrieving commit history
		for repo := range repoList {
			parsedURL, err := url.Parse(repo)
			if err != nil {
				log.Errorf("Skipping repo \"%s\" error parsing URL: %v\n", repo, err)
				continue
			}

			// Extract the host and switch between supported git remotes
			switch host := parsedURL.Host; host {
			case "github.com":
				// Extract github owner and repo from the parsed URL
				path := parsedURL.Path
				split := strings.Split(strings.TrimPrefix(path, "/"), "/")
				githubRepo(split[0], split[1])
			default:
				log.Errorf("Currently unsupported repository declared: %s", repo)
				continue
			}
		}

		// This interval wait is 60 minutes by default and specified with the interval flag.
		// We could tighten these intervals, though this tool makes a lot of API calls and we may run into rate limits from git remotes
		log.Debugf("waiting %d minutes before starting the next collector interval", interval)
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func createRepoList(cfg config.Config) error {
	// Move individual repo list to a map
	for _, repo := range cfg.IndividualRepositories {
		repoList[repo] = true
	}

	// Add organization repos to map
	for _, org := range cfg.GithubOrganizations {
		log.Debugf("adding repos from GitHub Organization %s with visibility %s to repo list", org.Name, org.Visibility)
		repos, err := gh.ListRepositoriesByOrg(org.Name, org.Visibility)
		if err != nil {
			return fmt.Errorf("error getting repository list by org for %s", org.Name)
		}

		for _, r := range repos {
			if org.ExcludeForks && *r.Fork {
				log.Infof("skipping %s (FORK)", *r.HTMLURL)
				continue
			}
			repoList[*r.HTMLURL] = true
		}
	}

	return nil
}

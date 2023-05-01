package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v52/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	client *github.Client
)

// Init constructs a github API client for this package taking in a token
func Init(token string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)
}

// GetRepository gets a repository by owner and repo name
func GetRepository(owner string, repo string) (*github.Repository, int, error) {
	log.Debugf("Querying GetRepository for %s/%s", owner, repo)
	// Get page of repo commits
	r, resp, err := client.Repositories.Get(context.Background(), owner, repo)
	var statusCode int
	if resp != nil {
		statusCode = resp.StatusCode
	}
	if err != nil {
		return nil, statusCode, fmt.Errorf("GetRepository for %s/%s returned error: \n%v", owner, repo, err)
	}

	log.Debugf("Queried GetRepository for %s/%s", owner, repo)
	return r, statusCode, nil
}

// ListRepositoryCommits gets all commits for a repository for a specified duration
func ListRepositoryCommits(owner string, repo string, start time.Time, end time.Time) ([]*github.RepositoryCommit, int, error) {
	var commits []*github.RepositoryCommit
	var statusCode int
	var page, perPage int = 1, 100
	for {
		data := github.CommitsListOptions{
			Since: start,
			Until: end,
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		log.Debugf("Querying ListRepositoryCommits for %s/%s, page %d", owner, repo, page)
		// Get page of repo commits
		r, resp, err := client.Repositories.ListCommits(context.Background(), owner, repo, &data)
		statusCode = resp.StatusCode
		if err != nil {
			return nil, statusCode, fmt.Errorf("ListRepositoriesByOrg returned error: \n%v", err)
		}

		// Add page to commits slice
		commits = append(commits, r...)

		// Break if out of pages, or flip page
		if resp.NextPage == 0 {
			break
		}
		page++
	}

	log.Debugf("Queried ListRepositoryCommits for %s/%s, with %d results", owner, repo, len(commits))
	return commits, statusCode, nil
}

// ListRepositoriesByOrg gets all repositories in a GitHub organization with a visibility filter setting
func ListRepositoriesByOrg(org string, visibility string) ([]*github.Repository, error) {
	// Chech cache
	var repos []*github.Repository
	var page, perPage int = 1, 100
	for {
		data := github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}
		if visibility != "" {
			data.Type = visibility
		}

		log.Debugf("Querying ListRepositoriesByOrg for %s of type (%s), page %d", org, visibility, page)
		// Get page of organization repos
		r, resp, err := client.Repositories.ListByOrg(context.Background(), org, &data)
		if err != nil {
			return nil, fmt.Errorf("ListRepositoriesByOrg returned error: \n%v", err)
		}

		// Add page to repos slice
		repos = append(repos, r...)

		// Break if out of pages, or flip page
		if resp.NextPage == 0 {
			break
		}
		page++
	}

	log.Debugf("Queried ListRepositoriesByOrg for %s of type %s, with %d results", org, visibility, len(repos))
	return repos, nil
}

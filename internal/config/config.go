package config

// Config Represents the application level config
type Config struct {
	GithubOrganizations    []GithubOrganizations `mapstructure:"github_organizations"`
	IndividualRepositories []string              `mapstructure:"individual_repositories"` // Individual repositories (not owned by specific orgs or users)
}

// GithubOrganizations represents key attributes for a github organization for this config
type GithubOrganizations struct {
	Name         string `mapstructure:"name"`          // The name of the org
	Visibility   string `mapstructure:"visibility"`    // The visibility level of repos to look at
	ExcludeForks bool   `mapstructure:"exclude_forks"` // Set to true if you want to exclude repo forks from the organization
}

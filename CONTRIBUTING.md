# Contributing

Thanks for your interest in contributing! The most common contribution to this repo is updating `config.yaml` to add repositories to the set of projects that this tool collects activity data for. This guide walks through how to do that, plus the one repo-wide requirement you should know about before opening a PR: signed commits.

## Creating Signed Commits

Our branch protection rules require that all commits be signed. If you haven't signed your commits before, you can read about commit signing here: https://docs.github.com/en/authentication/managing-commit-signature-verification

There are detailed, per-OS steps for setting up commit signing with either [GPG keys](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key#telling-git-about-your-gpg-key), [SSH keys](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key#telling-git-about-your-ssh-key), or [X.509 keys](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key#telling-git-about-your-x509-key).

The `🚨 Check commit signing` workflow runs on every pull request and will fail if any commit in your branch is unsigned, so it's worth getting this set up before you push. If you're unsure about setting up commit signing locally or would rather not, open up an Issue and someone else can add to the repository list on your behalf!

## Adding a repository

Please add your repository as an entry to `individual_repositories`. A few things to keep in mind:

- Use the full `https://github.com/<owner>/<repo>` URL. No trailing slash, no `.git` suffix.
- Double-check the spelling of the scheme and host (`https://github.com/`). A typo here means the entry will be silently skipped or fail to resolve.
- Try to keep the list reasonably ordered (the existing list is roughly alphabetical by owner) so that diffs stay easy to review.
- Avoid duplicates — search the file for the owner/repo before adding.
- Yaml requires consistent spacing - Begin a new listing with two spaces, a hyphen, another space, and then the repository URL.

Example:

```yaml
individual_repositories:
  - https://github.com/some-owner/some-repo
```

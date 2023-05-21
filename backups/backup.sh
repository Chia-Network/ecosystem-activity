#!/bin/bash

AWS_S3_BUCKET="s3://chia-ecosystem-github-repo-backups"
yaml_file="./config.yaml"
repositories=($(yq eval '.individual_repositories[]' "$yaml_file"))

# Enable case-insensitive matching
shopt -s nocasematch

# Loop through the repositories configured in the individual_repositories list
for repo in "${repositories[@]}"; do
  # Extract the organization and repository from the URL
  repo_url="${repo#https://github.com/}"
  organization="${repo_url%%/*}"
  repository="${repo_url#*/}"

  # Check if the organization is "chia-network" (case-insensitive)
  if [[ $organization == "chia-network" ]]; then
    echo "Organization is chia-network. Skipping iteration for ${organization}/${repository}."
    continue
  fi

  echo "Cloning repository ${repository}"
  clone_url="https://${GH_TOKEN}@github.com/${organization}/${repository}.git"
  git clone --mirror "${clone_url}"

  echo "Copying ${repository} to S3"
  tar czf "${repository}.git.tar.gz" "${repository}.git"
  aws s3 mv "${repository}.git.tar.gz" "${AWS_S3_BUCKET}/${repository}.git.tar.gz"
  rm -r "${repository}.git"

  echo "Checking for existence of a wiki repo for ${repository}"
  clone_url="https://${GH_TOKEN}@github.com/${organization}/${repository}.wiki.git"
  git clone --mirror "${clone_url}" || true # Ignore errors here if wiki does not exist
  if [ -d "${repository}.wiki.git" ]; then
	echo "Copying ${repository}.wiki to S3"
	tar czf "${repository}.wiki.git.tar.gz" "${repository}.wiki.git"
	aws s3 mv "${repository}.wiki.git.tar.gz" "${AWS_S3_BUCKET}/${repository}.wiki.git.tar.gz"
	rm -r "${repository}.wiki.git"
  fi
done

# ecosystem-activity

## Local dev

To run this locally, there's a `docker-compose.yml.tmpl` file at the root of the repo. It has the `.tmpl` extension for a reason, and it's to denote that there's some effort to customize it for your needs. But not a lot! That's what this little guide is for. Also this guide is assuming you have already cloned the repository and switched to the repo's directory.

First, copy the file over so we can make edits to it as needed without altering the template file used for this local dev guide.

```bash
cp docker-compose.yml.tmpl docker-compose.yml
```

Now open up this `docker-compose.yml` file and there are multiple environment variables in the collector service to change, and a volume for your config file.

```bash
environment:
  ECOSYSTEM_ACTIVITY_GITHUB_TOKEN: changeme
```

This environment variable in particular must be changed to your GitHub API token for the application to function.

Now we can create a testing configuration file. The configuration for the tool is loaded via CLI flags and environment variables. The configuration _file_ specifies all of the repositories that are in scope for data collection by this tool. For local dev, I would recommend creating a file at the root of this repo named `testconfig.yaml` and populating it with one or a few repositories that you want to collect test data from. Not too many, the point is to simulate production not _be_ production. This file, if named the same way, already has a volume mount specified in the docker-compose file for it, so nothing more should need to be changed. Here's an example minimal config with one repository selected:

```yaml
individual_repositories:
  - https://github.com/Chia-Network/go-chia-libs
```

At this point just run the following to start the application:

```bash
docker-compose up --build
```

You should see loglines flow in for the collector image to build. Then the mysql database container should start up. Once mysql is ready to accept connections the collector container should start and you'll see logs flow in at a debug level if you didn't change the `ECOSYSTEM_ACTIVITY_LOG_LEVEL` environment variable. If you make changes to the application code, just ctrl+c out of the docker-compose log stream, and re-run `docker-compose up --build` in your shell, and you're off to the races.

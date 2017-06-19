## Development Tasks

- [ ] Emit WebHook POST with Build Results
- [ ] Explore Huginn Integration platform: covert WebHook to Slack, Push, email, S3, etc. Need one-to-many notification.
- [X] ~~*Push Docker image to Docker Registry*~~
- [ ] Explore docker/cliconfig package for "reusing" already stored credentials or Docker config files
- [X] ~~*Decide on build status: Passed/Failed*~~
- [ ] Get BitBucket/GitLab webhooks (another URI path)
- [ ] Explore IFTTT and Zappier for easier integration (WebHook to Other) - write guide
- [ ] Remove Slack notification if flavour of webhook
- [ ] Support *private* repositories
- [ ] Pass all parameters to the `server` command through configuration file (as argument)
- [ ] Write Tests and Collect Coverage
- [ ] MicroCI with MicroCI
- [ ] Support build arguments (useful for secret passing during build: API keys, credentials)
- [ ] Support build tuneup with some property file in repo: define labels, environment, build arguments
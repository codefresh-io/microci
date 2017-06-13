# MicroCI

## What is MicroCI

*MicroCI* is minimalistic Continuous Integration (CI) tool, that implements a super simple CI pipeline for Docker-based microservices.

The Docker container technology simplifies build, deployment and operation of Microservices. 

*MicroCI* takes several assumptions about how development lifecycle of Docker-based microservices:

1. Docker image is a core package format and main deployment artifact
2. One git repository for one microservice
3. One Dockerfile to create and unit test microservice image, use multi-stage build for advanced built/test pipelines
4. Docker Registry (public or private) is used to store built Docker images
5. Service build should be triggered by Push, Branch or Tag event
6. Build report should be sent into Team group chat (Slack, or other)
7. Build report should be sent as WebHook payload for further integrations

### MicroCI Pipeline

1. **Trigger** - GitHub event: `push`, `branch`, `tag`
2. **Build** - `docker build` for Git "event" commit SHA
3. **Deploy** - `docker push` to Docker Registry (private, public)
4. **Notify** - send build report (status, duration, output, etc.) to STDOUT/Slack/WebHook

## Installation 

*MicroCI* is available as a single binary for every platform (Linux, Mac and Windows). It's statically linked without any external dependency. 

### MicroCI Docker container
*TBD*

### MicroCI Kubernetes Deployment
*TBD:*
## Connecting to Docker daemon

*MicroCI* can is working with Docker daemon through official Docker API (Go SDK). It can connect either to a local Docker daemon (no configuration is required in this case) or with a remote Docker daemon.

*MicroCI* respects all Docker environment variables when connecting to a Docker server. 

Use `DOCKER_HOST` to set the url to the docker server. Use `DOCKER_API_VERSION` to set the version of the API to reach, leave empty for latest. Use `DOCKER_CERT_PATH` to load the TLS certificates from. Use `DOCKER_TLS_VERIFY` to enable or disable TLS verification, off by default.
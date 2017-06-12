# MicroCI

## Docker Connection

*MicroCI* respects Docker environment variables when connecting to Docker server. 

Use `DOCKER_HOST` to set the url to the docker server. Use `DOCKER_API_VERSION` to set the version of the API to reach, leave empty for latest. Use `DOCKER_CERT_PATH` to load the TLS certificates from. Use `DOCKER_TLS_VERIFY` to enable or disable TLS verification, off by default.
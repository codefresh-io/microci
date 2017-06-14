#!/bin/sh
set -e

if [ "$1" = "microci" ]; then
  if [ -S /var/run/docker.sock ]; then
    chown -R microci:microci /var/run/docker.sock
  fi
  exec gosu microci:microci "$@"
fi

exec "$@"

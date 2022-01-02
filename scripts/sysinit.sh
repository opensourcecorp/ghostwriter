#!/usr/bin/env bash
set -euo pipefail

goversion='1.17.5'

# Using this as a temp fix for not having our own OCI image registry;
# this stuff should go in a Containerfile
apt-get update
apt-get install -y \
  ca-certificates \
  curl \
  make \
  tar

curl -fsSL -o /tmp/go"${goversion}".tar.gz https://golang.org/dl/go"${goversion}".linux-amd64.tar.gz
tar -C /usr/local -xzf /tmp/go"${goversion}".tar.gz
go version

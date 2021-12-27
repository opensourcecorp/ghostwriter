#!/usr/bin/env bash
set -euo pipefail

# Using this as a temp fix for not having our own OCI image registry;
# this stuff should go in a Containerfile
apt-get update
apt-get install -y \
  make \
  golang

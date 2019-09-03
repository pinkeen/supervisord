#!/usr/bin/env bash

set -euo pipefail

export GO_VERSION="$(go version)"
export GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
export BUILT_BY="$(whoami)@$(hostname)"

goreleaser --skip-publish --rm-dist --debug --snapshot 
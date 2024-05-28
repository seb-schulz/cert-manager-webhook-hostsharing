#!/bin/bash

set -xeu

${DOCKER_BIN} run --env RENOVATE_REPOSITORIES --env RENOVATE_ONBOARDING=false --env RENOVATE_FORCE --env LOG_LEVEL --env RENOVATE_TOKEN=${GITHUB_TOKEN} --volume /tmp:/tmp --rm ghcr.io/renovatebot/renovate:latest

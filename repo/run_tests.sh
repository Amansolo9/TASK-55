#!/usr/bin/env bash
set -euo pipefail

# Prevent Git Bash / MSYS on Windows from rewriting container paths like /app
# into C:/Program Files/Git/app before they reach dockerd.
export MSYS_NO_PATHCONV=1
export MSYS2_ARG_CONV_EXCL='*'

# Backend: Go unit + API tests in the official Go container.
docker run --rm \
  -v "$(pwd)":/app \
  -w /app \
  golang:1.26.1-alpine \
  sh -c "go test ./unit_tests/... -v && go test ./API_tests/... -v"

# Frontend: Vitest template/partial unit tests in the official Node container.
# Copy sources into /tmp inside the container so node_modules is created on a
# container-local path (not the bind mount, which may be noexec on CI runners
# and trigger "sh: vitest: Permission denied"). Invoke vitest through `node`
# to avoid relying on the bin shebang's execute bit entirely.
docker run --rm \
  -v "$(pwd)":/app:ro \
  node:22-alpine \
  sh -c "cp -r /app/frontend_tests /tmp/frontend_tests \
    && cp -r /app/views /tmp/views \
    && rm -rf /tmp/frontend_tests/node_modules \
    && cd /tmp/frontend_tests \
    && npm install --no-audit --no-fund --silent \
    && node ./node_modules/vitest/vitest.mjs run"

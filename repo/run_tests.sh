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
docker run --rm \
  -v "$(pwd)":/app \
  -w /app/frontend_tests \
  node:22-alpine \
  sh -c "npm install --no-audit --no-fund --silent && npm test"

#!/bin/bash
set -e

go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/lint/golint@latest
#!/bin/bash
set -e

go mod download || echo "No modules to download"
go mod tidy || echo "No go.mod yet"

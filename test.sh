#!/usr/bin/env bash

set -eu

go vet
go test  -count=1 ./...

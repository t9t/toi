#!/usr/bin/env bash

set -eu

go vet
go test -v -count=1 ./...

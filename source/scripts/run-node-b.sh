#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p bin
go build -o bin/node ./cmd/node
exec ./bin/node -id node-b -http 127.0.0.1:8082 -seed testdata/node-b-files.json

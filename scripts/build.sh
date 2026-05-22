#!/usr/bin/env bash
set -e

VERSION=${1:-"latest"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
OUTPUT_DIR="dist"
MODULE="github.com/anush-data-portfolio/MCPSync"
LDFLAGS="-s -w \
  -X ${MODULE}/cmd.version=${VERSION} \
  -X ${MODULE}/cmd.commit=${COMMIT} \
  -X ${MODULE}/cmd.buildTime=${BUILD_TIME}"

mkdir -p "$OUTPUT_DIR"

echo "Building MCPSync v$VERSION (commit=$COMMIT)..."

GOOS=darwin  GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/mcpsync-darwin-arm64" ./mcpsync
GOOS=darwin  GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/mcpsync-darwin-amd64" ./mcpsync
GOOS=linux   GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/mcpsync-linux-amd64" ./mcpsync
GOOS=linux   GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/mcpsync-linux-arm64" ./mcpsync
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/mcpsync-windows-amd64.exe" ./mcpsync

echo ""
echo "Built binaries:"
ls -lh "$OUTPUT_DIR"
echo ""
echo "To install locally:"
echo "  sudo cp $OUTPUT_DIR/mcpsync-\$(go env GOOS)-\$(go env GOARCH) /usr/local/bin/mcpsync"

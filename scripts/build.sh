#!/usr/bin/env bash
set -e

VERSION=${1:-"latest"}
OUTPUT_DIR="dist"

mkdir -p "$OUTPUT_DIR"

echo "Building MCPSync v$VERSION..."

GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w -X github.com/anush-data-portfolio/MCPSync/cmd.version=$VERSION" -o "$OUTPUT_DIR/mcpsync-darwin-arm64" .
GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w -X github.com/anush-data-portfolio/MCPSync/cmd.version=$VERSION" -o "$OUTPUT_DIR/mcpsync-darwin-amd64" .
GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w -X github.com/anush-data-portfolio/MCPSync/cmd.version=$VERSION" -o "$OUTPUT_DIR/mcpsync-linux-amd64" .
GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w -X github.com/anush-data-portfolio/MCPSync/cmd.version=$VERSION" -o "$OUTPUT_DIR/mcpsync-linux-arm64" .
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X github.com/anush-data-portfolio/MCPSync/cmd.version=$VERSION" -o "$OUTPUT_DIR/mcpsync-windows-amd64.exe" .

echo ""
echo "Built binaries:"
ls -lh "$OUTPUT_DIR"
echo ""
echo "To install locally:"
echo "  sudo cp $OUTPUT_DIR/mcpsync-\$(go env GOOS)-\$(go env GOARCH) /usr/local/bin/mcpsync"

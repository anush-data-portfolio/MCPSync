// Package pathutil provides shared path helpers for MCPSync.
package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath expands a leading "~/" to the user's home directory and
// normalises the result with filepath.Clean to prevent path traversal.
func ResolvePath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	return filepath.Clean(p)
}

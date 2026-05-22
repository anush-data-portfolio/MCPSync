package pathutil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anush-data-portfolio/MCPSync/internal/pathutil"
)

func TestResolvePath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	got := pathutil.ResolvePath("~/foo/bar")
	want := filepath.Join(home, "foo", "bar")
	if got != want {
		t.Errorf("ResolvePath(~/foo/bar) = %q, want %q", got, want)
	}
}

func TestResolvePath_NoTilde(t *testing.T) {
	got := pathutil.ResolvePath("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("ResolvePath(/absolute/path) = %q, want unchanged", got)
	}
}

func TestResolvePath_Clean(t *testing.T) {
	got := pathutil.ResolvePath("a/b/../c")
	if strings.Contains(got, "..") {
		t.Errorf("ResolvePath should clean traversal, got %q", got)
	}
}

func TestResolvePath_TildeNoSlash(t *testing.T) {
	p := "~username"
	got := pathutil.ResolvePath(p)
	if got != p {
		t.Errorf("ResolvePath(%q) = %q, want unchanged", p, got)
	}
}

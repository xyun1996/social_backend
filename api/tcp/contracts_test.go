package tcp_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var realtimeSurfaces = []string{
	"chat",
	"gateway",
}

func TestTCPContractsCoverRealtimeSurfaces(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir api/tcp failed: %v", err)
	}

	actual := make([]string, 0, len(realtimeSurfaces))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "README.md" {
			continue
		}
		actual = append(actual, strings.TrimSuffix(entry.Name(), ".md"))
	}
	slices.Sort(actual)

	expected := append([]string(nil), realtimeSurfaces...)
	slices.Sort(expected)

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected tcp contract set: actual=%v expected=%v", actual, expected)
	}
}

func TestTCPREADMEListsRealtimeSurfaces(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("ReadFile api/tcp/README.md failed: %v", err)
	}
	content := string(raw)

	for _, surface := range realtimeSurfaces {
		requireContains(t, "README.md", content, "- ["+surface+"]("+surface+".md)")
	}
}

func requireContains(t *testing.T, path string, content string, needle string) {
	t.Helper()
	if !strings.Contains(content, needle) {
		t.Fatalf("%s is missing required declaration %q", path, needle)
	}
}

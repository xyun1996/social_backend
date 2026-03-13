package http_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var controlPlaneSurfaces = []string{
	"chat",
	"gateway",
	"guild",
	"identity",
	"invite",
	"ops",
	"party",
	"presence",
	"social",
	"worker",
}

func TestHTTPContractsCoverControlPlaneSurfaces(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir api/http failed: %v", err)
	}

	actual := make([]string, 0, len(controlPlaneSurfaces))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "README.md" {
			continue
		}
		actual = append(actual, strings.TrimSuffix(entry.Name(), ".md"))
	}
	slices.Sort(actual)

	expected := append([]string(nil), controlPlaneSurfaces...)
	slices.Sort(expected)

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected http contract set: actual=%v expected=%v", actual, expected)
	}
}

func TestHTTPREADMEListsControlPlaneSurfaces(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("ReadFile api/http/README.md failed: %v", err)
	}
	content := string(raw)

	for _, surface := range controlPlaneSurfaces {
		requireContains(t, "README.md", content, "- ["+surface+"]("+surface+".md)")
	}
}

func requireContains(t *testing.T, path string, content string, needle string) {
	t.Helper()
	if !strings.Contains(content, needle) {
		t.Fatalf("%s is missing required declaration %q", path, needle)
	}
}

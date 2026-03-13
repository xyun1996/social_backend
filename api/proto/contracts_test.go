package proto_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var runtimeServices = []string{
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

func TestProtoContractsCoverRuntimeServices(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir api/proto failed: %v", err)
	}

	protoFiles := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".proto" {
			continue
		}
		protoFiles[strings.TrimSuffix(entry.Name(), ".proto")] = struct{}{}
	}

	expected := append([]string(nil), runtimeServices...)
	slices.Sort(expected)
	actual := make([]string, 0, len(protoFiles))
	for name := range protoFiles {
		actual = append(actual, name)
	}
	slices.Sort(actual)

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected proto service set: actual=%v expected=%v", actual, expected)
	}
}

func TestHTTPContractsCoverRuntimeServices(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("..", "http"))
	if err != nil {
		t.Fatalf("ReadDir api/http failed: %v", err)
	}

	httpFiles := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "README.md" {
			continue
		}
		httpFiles[strings.TrimSuffix(entry.Name(), ".md")] = struct{}{}
	}

	expected := append([]string(nil), runtimeServices...)
	slices.Sort(expected)
	actual := make([]string, 0, len(httpFiles))
	for name := range httpFiles {
		actual = append(actual, name)
	}
	slices.Sort(actual)

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected http service set: actual=%v expected=%v", actual, expected)
	}
}

func TestProtoFilesDeclareRequiredHeaders(t *testing.T) {
	t.Parallel()

	for _, service := range runtimeServices {
		service := service
		t.Run(service, func(t *testing.T) {
			t.Parallel()

			path := service + ".proto"
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile %s failed: %v", path, err)
			}

			content := string(raw)
			requireContains(t, path, content, `syntax = "proto3";`)
			requireContains(t, path, content, "package social_backend."+service+".v1;")
			requireContains(t, path, content, `option go_package = "github.com/xyun1996/social_backend/api/proto/`+service+`/v1;`+service+`v1";`)
		})
	}
}

func TestProtoREADMEListsRuntimeServices(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("ReadFile api/proto/README.md failed: %v", err)
	}
	content := string(raw)

	for _, service := range runtimeServices {
		requireContains(t, "README.md", content, "- `"+service+"`")
	}
}

func requireContains(t *testing.T, path string, content string, needle string) {
	t.Helper()
	if !strings.Contains(content, needle) {
		t.Fatalf("%s is missing required declaration %q", path, needle)
	}
}

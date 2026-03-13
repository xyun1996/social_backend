package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

var realtimeSurfaces = []string{
	"chat",
	"gateway",
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail("getwd failed: %v", err)
	}

	inventory, err := scanInventory(root)
	if err != nil {
		fail("scan contract inventory failed: %v", err)
	}

	printSummary(inventory)

	problems := validateInventory(inventory)
	if len(problems) == 0 {
		fmt.Println("Contract inventory checks passed.")
		return
	}

	for _, problem := range problems {
		fmt.Fprintf(os.Stderr, "contract inventory check failed: %s\n", problem)
	}
	os.Exit(1)
}

type inventory struct {
	HTTPFiles  []string
	ProtoFiles []string
	TCPFiles   []string
	HTTPIndex  string
	ProtoIndex string
	TCPIndex   string
}

func scanInventory(root string) (inventory, error) {
	httpFiles, httpIndex, err := scanMarkdownDir(filepath.Join(root, "api", "http"))
	if err != nil {
		return inventory{}, err
	}
	protoFiles, protoIndex, err := scanProtoDir(filepath.Join(root, "api", "proto"))
	if err != nil {
		return inventory{}, err
	}
	tcpFiles, tcpIndex, err := scanMarkdownDir(filepath.Join(root, "api", "tcp"))
	if err != nil {
		return inventory{}, err
	}

	return inventory{
		HTTPFiles:  httpFiles,
		ProtoFiles: protoFiles,
		TCPFiles:   tcpFiles,
		HTTPIndex:  httpIndex,
		ProtoIndex: protoIndex,
		TCPIndex:   tcpIndex,
	}, nil
}

func scanMarkdownDir(dir string) ([]string, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}

	files := make([]string, 0)
	indexRaw, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		return nil, "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || entry.Name() == "README.md" {
			continue
		}
		files = append(files, strings.TrimSuffix(entry.Name(), ".md"))
	}
	slices.Sort(files)
	return files, string(indexRaw), nil
}

func scanProtoDir(dir string) ([]string, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}

	files := make([]string, 0)
	indexRaw, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		return nil, "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".proto" {
			continue
		}
		files = append(files, strings.TrimSuffix(entry.Name(), ".proto"))
	}
	slices.Sort(files)
	return files, string(indexRaw), nil
}

func validateInventory(inv inventory) []string {
	problems := make([]string, 0)

	if !slices.Equal(inv.HTTPFiles, sortedCopy(controlPlaneSurfaces)) {
		problems = append(problems, fmt.Sprintf("unexpected HTTP contract set: actual=%v expected=%v", inv.HTTPFiles, sortedCopy(controlPlaneSurfaces)))
	}
	if !slices.Equal(inv.ProtoFiles, sortedCopy(controlPlaneSurfaces)) {
		problems = append(problems, fmt.Sprintf("unexpected proto contract set: actual=%v expected=%v", inv.ProtoFiles, sortedCopy(controlPlaneSurfaces)))
	}
	if !slices.Equal(inv.TCPFiles, sortedCopy(realtimeSurfaces)) {
		problems = append(problems, fmt.Sprintf("unexpected TCP contract set: actual=%v expected=%v", inv.TCPFiles, sortedCopy(realtimeSurfaces)))
	}

	for _, surface := range controlPlaneSurfaces {
		if !strings.Contains(inv.HTTPIndex, "- ["+surface+"]("+surface+".md)") {
			problems = append(problems, "api/http/README.md is missing surface "+surface)
		}
		if !strings.Contains(inv.ProtoIndex, "- `"+surface+"`") {
			problems = append(problems, "api/proto/README.md is missing service "+surface)
		}
	}
	for _, surface := range realtimeSurfaces {
		if !strings.Contains(inv.TCPIndex, "- ["+surface+"]("+surface+".md)") {
			problems = append(problems, "api/tcp/README.md is missing surface "+surface)
		}
	}

	return problems
}

func printSummary(inv inventory) {
	fmt.Printf("HTTP contracts:  %s\n", strings.Join(inv.HTTPFiles, ", "))
	fmt.Printf("Proto contracts: %s\n", strings.Join(inv.ProtoFiles, ", "))
	fmt.Printf("TCP contracts:   %s\n", strings.Join(inv.TCPFiles, ", "))
}

func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	slices.Sort(out)
	return out
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

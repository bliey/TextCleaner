package batch

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"textcleaner/internal/output"
)

// TestRunNewFolderPreservesStructure reproduces the real "drag a folder
// in → output to a chosen folder" scenario: a multi-file input with a
// sub-directory, run under Mode=Custom, verifying the directory structure
// is preserved and the files actually land on disk.
//
// Input source is the FOLDER (per the user's spec: "单个文件夹输入 →
// `<root>/<basename>/<内部目录结构>`"). The two flat file paths are
// only created on disk; what the batch driver receives is the folder.
func TestRunNewFolderPreservesStructure(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "src")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	f1 := filepath.Join(src, "a.txt")
	f2 := filepath.Join(src, "sub", "b.txt")
	os.WriteFile(f1, []byte("trim \n"), 0o644)
	os.WriteFile(f2, []byte("trim \n"), 0o644)
	out := filepath.Join(base, "out")

	opts := Options{
		Paths:          []string{src}, // folder input, NOT flat files
		Process:        sampleOptions(),
		Output:         output.Options{Mode: output.ModeCustom, CustomPath: out},
		MaxConcurrency: 2,
	}
	sum, err := Run(opts, context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Total != 2 || sum.Succeeded != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	// Folder → `<root>/<basename>` (here, "src"), structure preserved.
	for _, rel := range []string{"a.txt", filepath.Join("sub", "b.txt")} {
		target := filepath.Join(out, "src", rel)
		if _, statErr := os.Stat(target); statErr != nil {
			t.Fatalf("output missing at %s: %v", target, statErr)
		}
		got, _ := os.ReadFile(target)
		if string(got) != "trim\n" {
			t.Fatalf("content wrong at %s: %q", target, string(got))
		}
	}
	// 原文件不应被修改
	orig, _ := os.ReadFile(f1)
	if string(orig) != "trim \n" {
		t.Fatalf("original modified: %q", string(orig))
	}
}

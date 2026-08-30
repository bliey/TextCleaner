package batch

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"textcleaner/internal/model"
	"textcleaner/internal/output"
)

func sampleOptions() model.ProcessOptions {
	return model.ProcessOptions{
		BasicClean: model.BasicCleanOptions{
			TrimTrailingWhitespace: true,
		},
		Replace: []model.ReplaceRule{},
	}
}

// TestRunCustomMode verifies output.ModeCustom: a user-picked root
// receives every input file as <root>/<basename>, with the original
// file untouched.
func TestRunCustomMode(t *testing.T) {
	src := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")
	f1 := filepath.Join(src, "x.txt")
	os.WriteFile(f1, []byte("trim me \n"), 0o644)

	opts := Options{
		Paths:   []string{f1},
		Process: sampleOptions(),
		Output: output.Options{
			Mode:       output.ModeCustom,
			CustomPath: out,
		},
		MaxConcurrency: 1,
	}
	sum, err := Run(opts, context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Succeeded != 1 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	target := filepath.Join(out, "x.txt")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("output not written: %v", err)
	}
	if string(got) != "trim me\n" {
		t.Fatalf("output content wrong: %q", string(got))
	}
	// 原文件不应被修改
	orig, _ := os.ReadFile(f1)
	if string(orig) != "trim me \n" {
		t.Fatalf("original modified: %q", string(orig))
	}
}

// TestRunCreateSiblingMode verifies the default (output.ModeCreateSibling):
// input is a folder, output is a `<basename>_TCoutput` sibling that
// preserves the internal directory structure.
//
// sampleOptions only enables TrimTrailingWhitespace, so leading
// whitespace is preserved (and that's what we assert on here).
func TestRunCreateSiblingMode(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "novel")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("trim me \n"), 0o644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("trail   \n"), 0o644)

	opts := Options{
		Paths:   []string{src},
		Process: sampleOptions(),
		Output: output.Options{
			Mode: output.ModeCreateSibling,
		},
		MaxConcurrency: 1,
	}
	sum, err := Run(opts, context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Succeeded != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}

	sibling := filepath.Join(root, "novel_TCoutput")
	if _, err := os.Stat(sibling); err != nil {
		t.Fatalf("sibling dir not created: %v", err)
	}
	gotA, _ := os.ReadFile(filepath.Join(sibling, "a.txt"))
	if string(gotA) != "trim me\n" {
		t.Errorf("a.txt = %q, want trimmed trailing", string(gotA))
	}
	gotB, _ := os.ReadFile(filepath.Join(sibling, "sub", "b.txt"))
	if string(gotB) != "trail\n" {
		t.Errorf("sub/b.txt = %q, want trimmed trailing", string(gotB))
	}
	// Source untouched.
	origA, _ := os.ReadFile(filepath.Join(src, "a.txt"))
	if string(origA) != "trim me \n" {
		t.Errorf("original modified: %q", string(origA))
	}
}

// TestRunCancel: an already-cancelled context should short-circuit Run
// before any file I/O happens. Uses non-existent paths to avoid any
// filesystem side effects (especially Windows AV scans on temp files).
func TestRunCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(Options{
		Paths: []string{
			filepath.Join(t.TempDir(), "nope-a.txt"),
			filepath.Join(t.TempDir(), "nope-b.txt"),
		},
		Process:        sampleOptions(),
		MaxConcurrency: 1,
	}, ctx, nil)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}

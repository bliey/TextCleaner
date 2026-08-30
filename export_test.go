package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeRelativePath(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"file.txt", false},
		{"sub/file.txt", false},
		{"sub\\file.txt", false},
		{"../evil.txt", true},
		{"a/../../evil.txt", true},
		{"/abs/path.txt", true},
		{"C:/evil.txt", true},
		{"C:\\evil.txt", true},
		{"", true},
	}
	for _, c := range cases {
		_, err := sanitizeRelativePath(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("sanitizeRelativePath(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
		}
	}
}

func TestSaveFileBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	svc := &AppService{}
	payload := base64.StdEncoding.EncodeToString([]byte("你好，世界"))
	if err := svc.SaveFileBytes(path, payload); err != nil {
		t.Fatalf("SaveFileBytes: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "你好，世界" {
		t.Fatalf("content mismatch: %q", data)
	}
}

func TestExportFiles(t *testing.T) {
	dir := t.TempDir()
	svc := &AppService{}
	files := []ExportFile{
		{Path: "a.txt", Data: base64.StdEncoding.EncodeToString([]byte("AAA"))},
		{Path: "sub/b.txt", Data: base64.StdEncoding.EncodeToString([]byte("BBB"))},
		{Path: "sub/c.txt", Data: base64.StdEncoding.EncodeToString([]byte("CCC"))},
	}
	n, err := svc.ExportFiles(dir, files)
	if err != nil {
		t.Fatalf("ExportFiles: %v", err)
	}
	if n != 3 {
		t.Fatalf("written = %d, want 3", n)
	}
	for _, rel := range []string{"a.txt", filepath.Join("sub", "b.txt"), filepath.Join("sub", "c.txt")} {
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	// 已存在文件冲突 → 自动加 (1)
	n2, err := svc.ExportFiles(dir, []ExportFile{
		{Path: "a.txt", Data: base64.StdEncoding.EncodeToString([]byte("AAA2"))},
	})
	if err != nil {
		t.Fatalf("ExportFiles second: %v", err)
	}
	if n2 != 1 {
		t.Fatalf("second written = %d, want 1", n2)
	}
	if _, err := os.Stat(filepath.Join(dir, "a (1).txt")); err != nil {
		t.Errorf("expected a (1).txt after conflict: %v", err)
	}
}

func TestChooseDialogNoApp(t *testing.T) {
	svc := &AppService{} // app 为 nil，走防御分支
	if _, err := svc.ChooseDirectory(); err == nil {
		t.Fatal("ChooseDirectory: expected error when app not initialized")
	}
	if _, err := svc.ChooseSavePath("x.txt"); err == nil {
		t.Fatal("ChooseSavePath: expected error when app not initialized")
	}
}
func TestExportFilesRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	svc := &AppService{}
	_, err := svc.ExportFiles(dir, []ExportFile{
		{Path: "../escape.txt", Data: base64.StdEncoding.EncodeToString([]byte("x"))},
	})
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
	if !strings.Contains(err.Error(), "目录穿越") {
		t.Fatalf("unexpected error: %v", err)
	}
}

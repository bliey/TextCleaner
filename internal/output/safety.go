package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Normalize returns a comparable absolute, cleaned path string.
// On Windows this produces a backslash form with the same drive
// letter casing that filepath.Abs would yield; on POSIX it yields
// the canonical absolute form. Two paths that refer to the same
// location on disk compare equal after Normalize (modulo symlink
// resolution, which os.Stat-level calls elsewhere handle if
// needed).
//
// Used by IsInside, IsSamePath, and any comparison of two paths
// coming from different parts of the app (the user's selections,
// the scanner's results, the planner's candidates).
func Normalize(p string) string {
	if p == "" {
		return p
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		// Fall back to Clean — better than returning an empty path.
		return filepath.Clean(p)
	}
	return filepath.Clean(abs)
}

// IsSamePath returns true if a and b refer to the same filesystem
// location after Normalize. Case-insensitive on Windows (where
// "C:\foo" and "c:\FOO" are the same path); case-sensitive on POSIX.
func IsSamePath(a, b string) bool {
	a = Normalize(a)
	b = Normalize(b)
	if a == b {
		return true
	}
	return equalPathCaseInsensitiveOnWindows(a, b)
}

// IsInside returns true if `child` is the same path as `parent` or
// lives strictly inside `parent`. The comparison is done via
// filepath.Rel so it is robust to trailing-separator / case /
// platform differences.
//
// Used by the scanner to exclude the output directory from input
// recursion (otherwise the output dir would be re-scanned as input
// next time and produce nested _cleaned / 处理结果 folders).
func IsInside(child, parent string) bool {
	c := Normalize(child)
	p := Normalize(parent)
	if equalPathCaseInsensitiveOnWindows(c, p) {
		return true
	}
	rel, err := filepath.Rel(p, c)
	if err != nil {
		return false
	}
	// `..` at the start of Rel's result means child is NOT inside
	// parent; an empty result means child == parent (already handled).
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// equalPathCaseInsensitiveOnWindows returns true when a and b are
// equal under Windows filesystem semantics (case-insensitive). On
// other platforms it falls through to a case-sensitive equality
// check, which matches POSIX semantics.
func equalPathCaseInsensitiveOnWindows(a, b string) bool {
	if isWindows() {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// isWindows detects the running OS at package init time. The output
// package always builds the same way, so a one-time constant is fine.
var isWindows = func() bool {
	return filepath.Separator == '\\' && os.PathSeparator == ';'
}

// exists reports whether the given path exists (file or directory).
func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// isDir reports whether p exists and is a directory.
func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// UniqueDir returns a directory path that does not yet exist, by
// appending " (1)", " (2)", ... to `candidate` until a free name is
// found. The check uses exists() (which is sufficient for output
// planning — we are racing the user, not other goroutines; the
// MkdirAll call right after is what actually claims the path).
//
// Used by:
//   - create_sibling mode: "D:\小说_cleaned" already exists →
//     try "D:\小说_cleaned (1)".
//   - custom mode: when the user's chosen path already exists and
//     is non-empty, we still happily use it (the user picked it);
//     UniqueDir is only called for create_sibling.
func UniqueDir(candidate string) string {
	candidate = filepath.Clean(candidate)
	if !exists(candidate) {
		return candidate
	}
	parent := filepath.Dir(candidate)
	base := filepath.Base(candidate)
	for i := 1; i < 10000; i++ {
		try := filepath.Join(parent, fmt.Sprintf("%s (%d)", base, i))
		if !exists(try) {
			return try
		}
	}
	// Fallback: return the original (the caller will likely surface
	// an error from MkdirAll).
	return candidate
}

// UniqueSiblingFileName returns a file path (under `dir`) whose base
// matches `baseName` and does not collide with any file already
// listed in `usedNames`. The caller is responsible for adding the
// returned name to `usedNames` to avoid future collisions within the
// same planning pass.
//
// This is the file-level counterpart to UniqueDir. Used in custom
// mode for multiple input files where two of them share a basename
// (e.g. `D:\A.txt` and `E:\A.txt` → `处理结果\A.txt` and
// `处理结果\A (1).txt`).
func UniqueSiblingFileName(dir, baseName string, usedNames map[string]bool) string {
	if !usedNames[baseName] && !exists(filepath.Join(dir, baseName)) {
		return filepath.Join(dir, baseName)
	}
	ext := filepath.Ext(baseName)
	stem := strings.TrimSuffix(baseName, ext)
	for i := 1; i < 10000; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", stem, i, ext)
		full := filepath.Join(dir, candidate)
		if !usedNames[candidate] && !exists(full) {
			return full
		}
	}
	return filepath.Join(dir, baseName)
}

// SafeWrite writes data to path atomically: write to `<path>.tmp`,
// then rename. The rename either fully succeeds or fully fails —
// we never leave path half-written even if the process crashes
// mid-write. Parent directories are created as needed (0755).
//
// On Windows, os.Rename refuses to overwrite an existing target.
// We work around that by removing the target first, then renaming.
// This is safe because SafeWrite is only ever called from the
// batch driver on a freshly-planned output path (output.Planner
// guarantees the target doesn't exist) — the in-place overwrite
// case was removed from the product, so we never need to replace
// an existing original.
func SafeWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent dir %s: %w", dir, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		// Windows-specific retry: remove dst first, then rename.
		_ = os.Remove(path)
		if err2 := os.Rename(tmp, path); err2 != nil {
			return fmt.Errorf("rename tmp→dst (after remove): %w / %v", err, err2)
		}
	}
	return nil
}

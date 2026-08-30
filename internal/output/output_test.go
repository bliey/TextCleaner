package output

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ============================================================
// helpers
// ============================================================

// mkTempDir creates a fresh temp directory under the OS temp dir and
// returns its absolute path. Cleanup is registered via t.Cleanup so the
// temp dir is removed even on test failure.
func mkTempDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "textcleaner-output-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(d) })
	return d
}

// writeFile is a small helper that creates parent dirs as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// ============================================================
// safety.go: Normalize, IsInside, IsSamePath, UniqueDir,
//            UniqueSiblingFileName, SafeWrite
// ============================================================

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		// Normalize goes through filepath.Abs → filepath.Clean, so the
		// output is an absolute path. We assert that it's non-empty and
		// ends in the trailing path component (platform-portable).
		{"./a", filepath.Clean(mustAbs("./a"))},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// Idempotence: Normalize(Normalize(p)) == Normalize(p).
	p := Normalize(filepath.Join(mkTempDir(t), "foo", "bar"))
	if Normalize(p) != p {
		t.Errorf("Normalize not idempotent: %q vs %q", p, Normalize(p))
	}
}

func mustAbs(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	return abs
}

func TestIsInside(t *testing.T) {
	root := mkTempDir(t)
	child := filepath.Join(root, "a", "b")
	_ = os.MkdirAll(child, 0o755)
	cousin := filepath.Join(root, "sibling", "x")
	_ = os.MkdirAll(cousin, 0o755)

	cases := []struct {
		child, parent string
		want          bool
	}{
		{root, root, true},                // equal path
		{child, root, true},               // strictly inside
		{cousin, child, false},            // cousin dir, not inside
		{filepath.Dir(root), root, false}, // parent dir, not inside
	}
	for _, c := range cases {
		if got := IsInside(c.child, c.parent); got != c.want {
			t.Errorf("IsInside(%q, %q) = %v, want %v", c.child, c.parent, got, c.want)
		}
	}
}

func TestUniqueDir(t *testing.T) {
	root := mkTempDir(t)
	base := filepath.Join(root, "out")
	// Base doesn't exist → returned unchanged.
	if got := UniqueDir(base); got != base {
		t.Errorf("UniqueDir on missing = %q, want %q", got, base)
	}
	// Create base → next call appends " (1)".
	if err := os.Mkdir(base, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := UniqueDir(base); !strings.HasSuffix(got, " (1)") {
		t.Errorf("UniqueDir on existing = %q, want suffix \" (1)\"", got)
	}
}

func TestUniqueSiblingFileName(t *testing.T) {
	root := mkTempDir(t)
	used := map[string]bool{}
	// First call: returns the requested name.
	if got := UniqueSiblingFileName(root, "A.txt", used); got != filepath.Join(root, "A.txt") {
		t.Errorf("got %q", got)
	}
	// Pretend "A.txt" is now used.
	used["A.txt"] = true
	if got := UniqueSiblingFileName(root, "A.txt", used); !strings.HasSuffix(got, "A (1).txt") {
		t.Errorf("got %q, want suffix \"A (1).txt\"", got)
	}
}

func TestSafeWrite(t *testing.T) {
	root := mkTempDir(t)
	path := filepath.Join(root, "deep", "nested", "out.txt")
	if err := SafeWrite(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("SafeWrite: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
	// No leftover tmp file.
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Fatalf("leftover tmp file at %s", path+".tmp")
	}
}

// ============================================================
// planner.go: BuildPlan(...) and the two modes
// ============================================================

// --- Test 1: single folder + include subdirs + default (create_sibling) ---
func TestPlan_CreateSibling_SingleFolder(t *testing.T) {
	root := mkTempDir(t)
	src := filepath.Join(root, "novel")
	writeFile(t, filepath.Join(src, "a.txt"), "A")
	writeFile(t, filepath.Join(src, "sub", "deep", "b.txt"), "B")

	plan, err := BuildPlan([]InputSource{{Path: src, IsDir: true}}, Options{Mode: ModeCreateSibling})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.Mode != ModeCreateSibling {
		t.Errorf("mode = %q", plan.Mode)
	}
	if len(plan.OutputRoots) != 1 {
		t.Fatalf("OutputRoots = %v, want 1 root", plan.OutputRoots)
	}
	sibling := plan.OutputRoots[0]
	if filepath.Base(sibling) != "novel_TCoutput" {
		t.Errorf("sibling name = %q, want novel_TCoutput", filepath.Base(sibling))
	}
	if !isDir(sibling) {
		t.Errorf("sibling dir %s not created", sibling)
	}
	// Both files mapped, with preserved internal structure.
	if len(plan.Mappings) != 2 {
		t.Fatalf("Mappings = %v, want 2", plan.Mappings)
	}
	gotA := false
	gotB := false
	for _, m := range plan.Mappings {
		if filepath.Base(m.Src) == "a.txt" && m.Dst == filepath.Join(sibling, "a.txt") {
			gotA = true
		}
		if filepath.Base(m.Src) == "b.txt" && m.Dst == filepath.Join(sibling, "sub", "deep", "b.txt") {
			gotB = true
		}
	}
	if !gotA || !gotB {
		t.Errorf("missing a.txt or b.txt mapping (got %v)", plan.Mappings)
	}
}

func TestPlan_CreateSibling_SiblingAlreadyExists_AppendsNumber(t *testing.T) {
	root := mkTempDir(t)
	src := filepath.Join(root, "novel")
	writeFile(t, src+".txt", "x") // unrelated file
	_ = os.Mkdir(src, 0o755)
	_ = os.Mkdir(src+"_TCoutput", 0o755)     // sibling already exists
	_ = os.Mkdir(src+"_TCoutput (1)", 0o755) // second conflict

	plan, err := BuildPlan([]InputSource{{Path: src, IsDir: true}}, Options{Mode: ModeCreateSibling})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	sibling := plan.OutputRoots[0]
	// The planner must pick the next free slot (" (2)") — that's the
	// semantic the user's spec requires. We don't assert isDir after
	// the call because Plan itself created the dir (we want to verify
	// the *name*, not the existence).
	if filepath.Base(sibling) != "novel_TCoutput (2)" {
		t.Errorf("sibling name = %q, want novel_TCoutput (2)", filepath.Base(sibling))
	}
	// Belt-and-suspenders: that name must not be one of the pre-existing
	// siblings that we set up at the top.
	for _, taken := range []string{src + "_TCoutput", src + "_TCoutput (1)"} {
		if Normalize(sibling) == Normalize(taken) {
			t.Errorf("planner picked a name that collides with existing dir %s", taken)
		}
	}
}

// --- Test 2: multiple input folders + custom output ---
func TestPlan_Custom_MultipleFolders(t *testing.T) {
	root := mkTempDir(t)
	novel := filepath.Join(root, "novel")
	articles := filepath.Join(root, "articles")
	writeFile(t, filepath.Join(novel, "a.txt"), "A")
	writeFile(t, filepath.Join(novel, "sub", "b.txt"), "B")
	writeFile(t, filepath.Join(articles, "c.txt"), "C")

	out := filepath.Join(root, "处理结果")
	plan, err := BuildPlan(
		[]InputSource{
			{Path: novel, IsDir: true},
			{Path: articles, IsDir: true},
		},
		Options{Mode: ModeCustom, CustomPath: out},
	)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.OutputRoot != out {
		t.Errorf("OutputRoot = %q, want %q", plan.OutputRoot, out)
	}
	// Expect 3 mappings: a.txt, sub/b.txt, c.txt
	if len(plan.Mappings) != 3 {
		t.Fatalf("Mappings = %d, want 3: %+v", len(plan.Mappings), plan.Mappings)
	}
	wantNovelA := filepath.Join(out, "novel", "a.txt")
	wantNovelB := filepath.Join(out, "novel", "sub", "b.txt")
	wantArticlesC := filepath.Join(out, "articles", "c.txt")
	got := map[string]string{}
	for _, m := range plan.Mappings {
		got[m.Src] = m.Dst
	}
	if got[filepath.Join(novel, "a.txt")] != wantNovelA {
		t.Errorf("a.txt dst = %q, want %q", got[filepath.Join(novel, "a.txt")], wantNovelA)
	}
	if got[filepath.Join(novel, "sub", "b.txt")] != wantNovelB {
		t.Errorf("sub/b.txt dst = %q, want %q", got[filepath.Join(novel, "sub", "b.txt")], wantNovelB)
	}
	if got[filepath.Join(articles, "c.txt")] != wantArticlesC {
		t.Errorf("articles/c.txt dst = %q, want %q", got[filepath.Join(articles, "c.txt")], wantArticlesC)
	}
}

func TestPlan_Custom_SingleFile(t *testing.T) {
	root := mkTempDir(t)
	a := filepath.Join(root, "D", "a.txt")
	writeFile(t, a, "A")
	out := filepath.Join(root, "处理结果")

	plan, err := BuildPlan(
		[]InputSource{{Path: a, IsDir: false}},
		Options{Mode: ModeCustom, CustomPath: out},
	)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(plan.Mappings) != 1 {
		t.Fatalf("Mappings = %d, want 1", len(plan.Mappings))
	}
	if plan.Mappings[0].Dst != filepath.Join(out, "a.txt") {
		t.Errorf("dst = %q", plan.Mappings[0].Dst)
	}
}

// --- Test 3: multiple independent files + filename conflict ---
func TestPlan_Custom_FilenameConflict_AutoNumbers(t *testing.T) {
	root := mkTempDir(t)
	a := filepath.Join(root, "D", "A.txt")
	b := filepath.Join(root, "E", "A.txt")
	c := filepath.Join(root, "F", "A.txt")
	writeFile(t, a, "1")
	writeFile(t, b, "2")
	writeFile(t, c, "3")
	out := filepath.Join(root, "处理结果")

	plan, err := BuildPlan(
		[]InputSource{
			{Path: a, IsDir: false},
			{Path: b, IsDir: false},
			{Path: c, IsDir: false},
		},
		Options{Mode: ModeCustom, CustomPath: out},
	)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	dstSet := map[string]bool{}
	for _, m := range plan.Mappings {
		dstSet[filepath.Base(m.Dst)] = true
	}
	for _, name := range []string{"A.txt", "A (1).txt", "A (2).txt"} {
		if !dstSet[name] {
			t.Errorf("missing destination %q; got %v", name, dstSet)
		}
	}
}

// --- Test 4: output directory already exists ---
func TestPlan_Custom_OutputRootAlreadyExists_ReusedAsIs(t *testing.T) {
	root := mkTempDir(t)
	out := filepath.Join(root, "处理结果")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	// Pre-existing file under out that should NOT be clobbered.
	preExisting := filepath.Join(out, "unrelated.txt")
	writeFile(t, preExisting, "keep me")

	src := filepath.Join(root, "novel", "a.txt")
	writeFile(t, src, "A")

	plan, err := BuildPlan(
		[]InputSource{{Path: src, IsDir: false}},
		Options{Mode: ModeCustom, CustomPath: out},
	)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.OutputRoot != out {
		t.Errorf("OutputRoot = %q, want %q", plan.OutputRoot, out)
	}
	// The pre-existing file must still be there (we don't delete).
	got, _ := os.ReadFile(preExisting)
	if string(got) != "keep me" {
		t.Errorf("pre-existing file overwritten: got %q", got)
	}
}

// --- Test 5: output dir inside input dir + include subdirs → scanner
// must exclude output dir. We don't have scanner here, but we can
// verify the planner refuses the dangerous configuration and the
// helper used by the scanner correctly excludes it. ---
func TestPlan_Custom_OutputInsideInput_Rejected(t *testing.T) {
	root := mkTempDir(t)
	src := filepath.Join(root, "novel")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(src, "处理结果")
	_, err := BuildPlan(
		[]InputSource{{Path: src, IsDir: true}},
		Options{Mode: ModeCustom, CustomPath: out},
	)
	if err == nil {
		t.Fatal("Plan accepted output dir inside input; expected ErrOutputInsideInput")
	}
	if !isErr(err, ErrOutputInsideInput) {
		t.Errorf("err = %v, want ErrOutputInsideInput", err)
	}
}

func TestPlan_Custom_OutputEqualsInput_Rejected(t *testing.T) {
	root := mkTempDir(t)
	src := filepath.Join(root, "novel")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := BuildPlan(
		[]InputSource{{Path: src, IsDir: true}},
		Options{Mode: ModeCustom, CustomPath: src},
	)
	if err == nil {
		t.Fatal("Plan accepted output == input; expected ErrOutputEqualsInput")
	}
	if !isErr(err, ErrOutputEqualsInput) {
		t.Errorf("err = %v, want ErrOutputEqualsInput", err)
	}
}

// isErr is a small helper that returns true if err is err or wraps it.
func isErr(err, target error) bool {
	if err == nil {
		return false
	}
	return err == target || strings.Contains(err.Error(), target.Error())
}

// ============================================================
// Edge: normalization handles platform-specific separators
// ============================================================

func TestNormalize_PlatformIndependentCompare(t *testing.T) {
	// On Windows, two paths differing only by separator form should
	// be detected as the same. We don't rely on the actual separator
	// — both forms go through filepath.Clean which produces the
	// native separator.
	if runtime.GOOS == "windows" {
		// "C:/foo" and "C:\foo" should compare equal.
		a := Normalize(`C:/foo/bar`)
		b := Normalize(`C:\foo\bar`)
		if !strings.EqualFold(a, b) {
			t.Errorf("Normalize on Windows: %q vs %q", a, b)
		}
	}
}

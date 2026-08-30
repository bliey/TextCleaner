package output

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BuildPlan builds an output Plan from the user's input sources + Options.
// It is the single entry point used by the batch driver; the per-mode
// implementations (planCreateSibling / planCustom) are internal helpers.
//
// Errors:
//   - mode is empty or unknown → ErrUnknownMode
//   - custom mode with empty CustomPath → ErrEmptyCustomPath
//   - custom mode where CustomPath is the same path as one of the
//     input sources → ErrOutputEqualsInput
//   - custom mode where CustomPath is inside one of the input sources
//     → ErrOutputInsideInput
//
// Note: the package also exports a `Plan` type (the result struct);
// the constructor is named `BuildPlan` to avoid Go's "type and
// function with the same name in the same scope" error.
func BuildPlan(inputs []InputSource, opt Options) (*Plan, error) {
	// Normalize + validate input sources up front.
	for i := range inputs {
		inputs[i].Path = Normalize(inputs[i].Path)
	}

	switch opt.Mode {
	case "", ModeCreateSibling:
		return planCreateSibling(inputs)
	case ModeCustom:
		if strings.TrimSpace(opt.CustomPath) == "" {
			return nil, ErrEmptyCustomPath
		}
		return planCustom(inputs, opt)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownMode, opt.Mode)
	}
}

// --- Sentinel errors (returned wrapped, callers can errors.Is) ---

var (
	ErrUnknownMode       = errors.New("output: unknown mode")
	ErrEmptyCustomPath   = errors.New("output: custom mode requires a non-empty CustomPath")
	ErrOutputEqualsInput = errors.New("output: custom path equals an input source")
	ErrOutputInsideInput = errors.New("output: custom path is inside an input source")
)

// ============================================================
// Mode 1: create_sibling (default)
//
// Input: D:\小说 (folder)
// Output: D:\小说_TCoutput\  (sibling folder, contains the file tree)
//
// Input: D:\A.txt (file)
// Output: D:\A_TCoutput\A.txt  (sibling folder named after the file stem)
//
// If the sibling folder already exists, we append " (1)", " (2)", ...
// until a free name is found. Files within a folder preserve their
// relative directory structure.
func planCreateSibling(inputs []InputSource) (*Plan, error) {
	plan := &Plan{Mode: ModeCreateSibling}

	for _, src := range inputs {
		// Anchor directory is the parent of the input source.
		// The sibling name is the input basename (or stem for files)
		// with "_TCoutput" appended.
		anchorDir := filepath.Dir(src.Path)
		var stem string
		if src.IsDir {
			stem = filepath.Base(src.Path)
		} else {
			stem = strings.TrimSuffix(filepath.Base(src.Path), filepath.Ext(src.Path))
		}
		siblingCandidate := filepath.Join(anchorDir, stem+"_TCoutput")
		sibling := UniqueDir(siblingCandidate)
		if !exists(sibling) {
			if err := os.MkdirAll(sibling, 0o755); err != nil {
				return nil, fmt.Errorf("create sibling dir %s: %w", sibling, err)
			}
		}
		plan.AddOutputRoot(sibling)

		if src.IsDir {
			// Walk the input tree; map each file to <sibling>/<relpath>.
			err := filepath.WalkDir(src.Path, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				rel, relErr := filepath.Rel(src.Path, path)
				if relErr != nil {
					return relErr
				}
				dst := filepath.Join(sibling, rel)
				plan.AddMapping(Mapping{Src: path, Dst: dst})
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk input %s: %w", src.Path, err)
			}
		} else {
			// Single file → placed inside the sibling folder under
			// its original basename.
			dst := filepath.Join(sibling, filepath.Base(src.Path))
			plan.AddMapping(Mapping{Src: src.Path, Dst: dst})
		}
	}
	return plan, nil
}

// ============================================================
// Mode 2: custom
//
// User-picked root (e.g. D:\处理结果). Behaviour depends on whether
// each input is a file or a directory:
//
//   - File:  D:\A.txt  →  D:\处理结果\A.txt
//   - Folder: D:\小说  →  D:\处理结果\小说\<file tree preserved>
//
// Conflicts at the root level (multiple inputs producing the same
// basename) are resolved by appending " (1)", " (2)", ...
//
// The root itself may already exist with content; we do NOT delete
// or rename it. We only ensure the file we're about to write won't
// silently clobber an existing file at the same path.
func planCustom(inputs []InputSource, opt Options) (*Plan, error) {
	root := Normalize(opt.CustomPath)
	if root == "" {
		return nil, ErrEmptyCustomPath
	}
	// Safety: output must not be inside or equal to any input.
	for _, src := range inputs {
		if IsSamePath(root, src.Path) {
			return nil, fmt.Errorf("%w: %s", ErrOutputEqualsInput, root)
		}
		if IsInside(root, src.Path) {
			return nil, fmt.Errorf("%w: %s inside %s", ErrOutputInsideInput, root, src.Path)
		}
	}
	// Ensure root exists. We never delete / overwrite an existing root
	// — the user picked this location explicitly.
	if !exists(root) {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return nil, fmt.Errorf("create custom root %s: %w", root, err)
		}
	}

	plan := &Plan{
		Mode:       ModeCustom,
		OutputRoot: root,
	}
	plan.AddOutputRoot(root)

	// Track basenames already emitted at the root level (files only),
	// to resolve filename conflicts per the user's spec.
	usedNames := map[string]bool{}

	for _, src := range inputs {
		if src.IsDir {
			// Folder → <root>/<basename>/<file tree>
			subRoot := filepath.Join(root, filepath.Base(src.Path))
			// Safety: if the user picked an output root that already
			// contains a subdirectory matching this input's name,
			// reuse it (don't overwrite unrelated content).
			plan.AddOutputRoot(subRoot)

			err := filepath.WalkDir(src.Path, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				rel, relErr := filepath.Rel(src.Path, path)
				if relErr != nil {
					return relErr
				}
				dst := filepath.Join(subRoot, rel)
				plan.AddMapping(Mapping{Src: path, Dst: dst})
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk input %s: %w", src.Path, err)
			}
		} else {
			// File → <root>/<basename> with conflict resolution
			base := filepath.Base(src.Path)
			dst := UniqueSiblingFileName(root, base, usedNames)
			usedNames[filepath.Base(dst)] = true
			plan.AddMapping(Mapping{Src: src.Path, Dst: dst})
		}
	}
	return plan, nil
}

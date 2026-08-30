// Package output implements the two output modes (create_sibling /
// custom), the path-planning logic that maps each input file to
// its output path, and the safety primitives (path normalization,
// conflict resolution, atomic write).
//
// Note on the missing third mode: 之前有「覆盖原文件」+ 自动 .backup
// 的第三条路径，在用户要求下已下线 —— 默认安全第一，不覆盖原文件。
// 所有输出都写到新位置。
//
// Architectural goals (per project spec):
//
//   - Default safety: never silently overwrite the original file.
//     Every output lands in a fresh sibling / user-picked dir.
//   - No magic numbers: Mode is a string enum; Options carries the
//     user's choices.
//   - Path logic is independent: no text-processing code anywhere
//     else in the repo should compute output paths.
//   - All path comparisons go through Normalize + IsInside so the
//     Windows / POSIX / case / symlink edge cases are handled in
//     one place.
package output

import (
	"path/filepath"
)

// Mode is the output strategy chosen by the user. String-typed so it
// can be persisted (settings.json) and bound to TypeScript without a
// separate enum mapping.
type Mode string

const (
	// ModeCreateSibling: 默认。安全、不覆盖原文件：
	// - 输入为文件夹 → 在其同级创建 `<basename>_TCoutput`
	// - 输入为单文件 → 在其父目录下创建 `<stem>_TCoutput`，文件落入该文件夹
	// 冲突时自动加 " (1)" / " (2)" 直至找到空名。
	ModeCreateSibling Mode = "create_sibling"

	// ModeCustom: 用户指定一个根输出目录：
	// - 文件夹输入 → `<root>/<basename>/<内部目录结构>`
	// - 文件输入   → `<root>/<basename>`（顶层冲突自动加 " (1)"）
	// 若 root 位于某个输入内部、被某个输入占用或为某个输入，BuildPlan 拒绝并报错。
	ModeCustom Mode = "custom"
)

// Options is the user-facing output configuration passed by the
// frontend / persisted across sessions.
//
// 跨语言 binding：JSON tag 即前端字段名（Wails 自动生成同名 TS 字段）。
type Options struct {
	Mode Mode `json:"mode"`

	// CustomPath is the user-picked root output directory, only used
	// when Mode == ModeCustom. Ignored for ModeCreateSibling.
	CustomPath string `json:"customPath"`
}

// InputSource is one of the items the user added to the batch (a file
// or a directory). Whether it's a dir determines how the planner
// lays out the output.
type InputSource struct {
	Path  string
	IsDir bool
}

// Mapping is a single (src, dst) pair that the batch worker should
// process: read src, run ProcessText, write to dst.
type Mapping struct {
	Src string // absolute path of the input file
	Dst string // absolute path of the output file
}

// Plan is the resolved output plan for a batch run. After the planner
// returns, every Dst is guaranteed to point to a path that does not
// collide with existing files (conflict resolution was applied during
// planning). Mappings is the list of (src, dst) pairs the worker
// should process.
//
// OutputRoot for each mode:
//   - create_sibling: each input source has its own `_TCoutput` sibling;
//     OutputRoot is empty here (use per-source dirs from Mappings).
//   - custom: OutputRoot is the user's chosen path.
type Plan struct {
	Mode        Mode
	OutputRoot  string   // "" when mode-specific (see above)
	Mappings    []Mapping
	OutputRoots []string // every distinct dir receiving outputs (for scanner exclusion)
}

// AddMapping deduplicates by Dst: if a Mapping with the same Dst is
// already in the slice, it is replaced. This protects the planner
// from emitting two writes to the same output path (which would cause
// a race or silent clobbering).
func (p *Plan) AddMapping(m Mapping) {
	for i, existing := range p.Mappings {
		if existing.Dst == m.Dst {
			p.Mappings[i] = m
			return
		}
	}
	p.Mappings = append(p.Mappings, m)
}

// AddOutputRoot adds dir to OutputRoots if it is not already present.
func (p *Plan) AddOutputRoot(dir string) {
	dir = filepath.Clean(dir)
	for _, existing := range p.OutputRoots {
		if existing == dir {
			return
		}
	}
	p.OutputRoots = append(p.OutputRoots, dir)
}

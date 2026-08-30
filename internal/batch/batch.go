package batch

import (
	"context"
	"os"
	"path/filepath"

	"textcleaner/internal/encoding"
	"textcleaner/internal/model"
	"textcleaner/internal/output"
	"textcleaner/internal/processor"
)

// FileResult 单个文件的处理结果。
type FileResult struct {
	Path     string `json:"path"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Deleted  int    `json:"deleted"`
	Replaced int    `json:"replaced"`
	Changed  bool   `json:"changed"`
}

// Summary 批量处理的汇总。
type Summary struct {
	Total     int          `json:"total"`
	Processed int          `json:"processed"`
	Succeeded int          `json:"succeeded"`
	Failed    int          `json:"failed"`
	Results   []FileResult `json:"results"`
}

// Options 批量处理选项。
//
// 输出配置由 internal/output.BuildPlan 统一处理（create_sibling /
// custom 两种模式），原覆盖模式已下线。
type Options struct {
	Paths          []string             `json:"paths"`
	Process        model.ProcessOptions `json:"process"`
	Output         output.Options       `json:"output"`
	OutputEncoding string               `json:"outputEncoding"`
	MaxConcurrency int                  `json:"maxConcurrency"`
}

// Progress 进度回调载荷（通过事件推送到前端）。
type Progress struct {
	Done    int        `json:"done"`
	Total   int        `json:"total"`
	Current string     `json:"current"`
	Last    FileResult `json:"last"`
}

// Run 以受限并发处理多个文件。onProgress 在每个文件完成后被调用（可能并发）。
//
// 整体流程：
//  1. 解析每个输入路径（确定 IsDir），构造 output.InputSource 列表；
//  2. 调用 output.BuildPlan 得到完整输出计划（含冲突解决 + 路径安全）；
//  3. 过滤掉任何落在输出目录内部的源文件（防御性，正常不会发生）；
//  4. 并发处理 plan.Mappings：读 src → ProcessText → SafeWrite 到 dst。
func Run(opts Options, ctx context.Context, onProgress func(Progress)) (Summary, error) {
	// ① 解析输入：每个路径都要确认是文件还是目录
	sources, err := resolveSources(opts.Paths)
	if err != nil {
		return Summary{}, err
	}
	if len(sources) == 0 {
		return Summary{}, nil
	}

	// ② 由 output 包生成完整输出计划
	plan, err := output.BuildPlan(sources, opts.Output)
	if err != nil {
		return Summary{}, err
	}

	// ③ 过滤：任何 Src 落在任一 OutputRoot 内的 Mapping 一律跳过
	// （安全网；正常情况下 planner 不会产生这种 mapping）
	validMappings := make([]output.Mapping, 0, len(plan.Mappings))
	for _, m := range plan.Mappings {
		skip := false
		for _, root := range plan.OutputRoots {
			if root == "" {
				continue
			}
			if output.IsInside(m.Src, root) {
				skip = true
				break
			}
		}
		if !skip {
			validMappings = append(validMappings, m)
		}
	}

	summary := Summary{
		Total:   len(validMappings),
		Results: make([]FileResult, 0, len(validMappings)),
	}
	if len(validMappings) == 0 {
		return summary, nil
	}

	// ④ 并发 worker 池
	conc := opts.MaxConcurrency
	if conc <= 0 {
		conc = 4
	}
	sem := make(chan struct{}, conc)

	type indexed struct {
		i int
		r FileResult
	}
	done := make(chan indexed, len(validMappings))

	for i, m := range validMappings {
		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		case sem <- struct{}{}:
		}

		go func(idx int, m output.Mapping) {
			defer func() { <-sem }() // 释放并发槽
			res := processMapping(m, opts)
			done <- indexed{i: idx, r: res}
		}(i, m)
	}

	// 收集结果
	results := make([]FileResult, len(validMappings))
	processed := 0
	succeeded := 0
	failed := 0
	for range validMappings {
		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		case x := <-done:
			results[x.i] = x.r
			processed++
			if x.r.Success {
				succeeded++
			} else {
				failed++
			}
			if onProgress != nil {
				onProgress(Progress{
					Done:    processed,
					Total:   len(validMappings),
					Current: x.r.Path,
					Last:    x.r,
				})
			}
		}
	}
	summary.Processed = processed
	summary.Succeeded = succeeded
	summary.Failed = failed
	summary.Results = results
	return summary, nil
}

// resolveSources 把前端传入的混合 file/dir 路径列表解析成
// output.InputSource。返回 (sources, error)，其中 error 仅在路径根本
// 不存在或不可访问时出现；空路径会被跳过。
func resolveSources(paths []string) ([]output.InputSource, error) {
	out := make([]output.InputSource, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, err
		}
		out = append(out, output.InputSource{
			Path:  abs,
			IsDir: info.IsDir(),
		})
	}
	return out, nil
}

// processMapping 是单文件处理流水线：
//
//	读 src → 解码 → ProcessText → 编码 → 原子写到 dst
func processMapping(m output.Mapping, opts Options) FileResult {
	data, err := os.ReadFile(m.Src)
	if err != nil {
		return FileResult{Path: m.Src, Success: false, Error: err.Error()}
	}

	text, srcEnc, _ := encoding.Decode(data)
	out, res, err := processor.ProcessText(text, opts.Process)
	if err != nil {
		return FileResult{Path: m.Src, Success: false, Error: err.Error()}
	}
	outBytes := encoding.Encode(out, encoding.OutputEncoding(opts.OutputEncoding), srcEnc)

	// create_sibling / custom 模式统一写到 planner 算好的 dst；
	// 覆盖模式已下线。
	if writeErr := output.SafeWrite(m.Dst, outBytes, 0o644); writeErr != nil {
		return FileResult{Path: m.Src, Success: false, Error: writeErr.Error()}
	}

	return FileResult{
		Path:     m.Src,
		Success:  true,
		Deleted:  res.DeletedMatches,
		Replaced: res.ReplacedMatches,
		Changed:  res.Changed,
	}
}

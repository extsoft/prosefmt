package rules

import (
	"bytes"
	"fmt"
)

const (
	PF6ID = "PF6"
	PF7ID = "PF7"
)

func PF6Msg(n int) string {
	return fmt.Sprintf("tab characters must be replaced with %d spaces", n)
}

func PF7Msg(n int) string {
	return fmt.Sprintf("each run of %d space characters must be replaced with a tab character", n)
}

func CheckPF6(file string, content []byte, tabWidth int) []Issue {
	if tabWidth <= 0 {
		return nil
	}
	var issues []Issue
	lines := splitLines(content)
	for lineNum, raw := range lines {
		contentPart, _ := stripLineEnding(raw)
		for i, b := range contentPart {
			if b == '\t' {
				issues = append(issues, Issue{
					File:    file,
					Line:    lineNum + 1,
					Column:  i + 1,
					RuleID:  PF6ID,
					Message: PF6Msg(tabWidth),
				})
				break
			}
		}
	}
	return issues
}

func FixTabs(content []byte, tabWidth int) []byte {
	if tabWidth <= 0 {
		return content
	}
	rep := bytes.Repeat([]byte{' '}, tabWidth)
	return bytes.ReplaceAll(content, []byte{'\t'}, rep)
}

func CheckPF7(file string, content []byte, runLen int) []Issue {
	if runLen <= 0 {
		return nil
	}
	var issues []Issue
	lines := splitLines(content)
	for lineNum, raw := range lines {
		contentPart, _ := stripLineEnding(raw)
		idx := indexOfFirstSpaceRun(contentPart, runLen)
		if idx < 0 {
			continue
		}
		issues = append(issues, Issue{
			File:    file,
			Line:    lineNum + 1,
			Column:  idx + 1,
			RuleID:  PF7ID,
			Message: PF7Msg(runLen),
		})
	}
	return issues
}

func indexOfFirstSpaceRun(b []byte, n int) int {
	if n <= 0 || len(b) < n {
		return -1
	}
	for i := 0; i+n <= len(b); i++ {
		if b[i] != ' ' {
			continue
		}
		ok := true
		for j := 1; j < n; j++ {
			if b[i+j] != ' ' {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}

func FixSpacesToTab(content []byte, runLen int) []byte {
	if runLen <= 0 {
		return content
	}
	pat := bytes.Repeat([]byte{' '}, runLen)
	for bytes.Contains(content, pat) {
		content = bytes.ReplaceAll(content, pat, []byte{'\t'})
	}
	return content
}

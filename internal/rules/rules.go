package rules

import "os"

const (
	PF4ID  = "PF4"
	PF4Msg = "line endings must be LF (linux-style)"
	PF5ID  = "PF5"
	PF5Msg = "line endings must be CRLF (windows-style)"
)

func Check(file string, content []byte, mode LineEndingMode, tabWidth int, spacesToTab int) []Issue {
	var issues []Issue
	issues = append(issues, CheckPF2(file, content)...)
	if tabWidth > 0 {
		issues = append(issues, CheckPF6(file, content, tabWidth)...)
	}
	if spacesToTab > 0 {
		issues = append(issues, CheckPF7(file, content, spacesToTab)...)
	}
	issues = append(issues, CheckPF1(file, content)...)
	for _, line := range lineEndingMismatchLines(content, mode) {
		if mode == LineEndLinux {
			issues = append(issues, Issue{File: file, Line: line, Column: 1, RuleID: PF4ID, Message: PF4Msg})
		} else if mode == LineEndWindows {
			issues = append(issues, Issue{File: file, Line: line, Column: 1, RuleID: PF5ID, Message: PF5Msg})
		}
	}
	return issues
}

func Fix(content []byte, mode LineEndingMode, tabWidth int, spacesToTab int) []byte {
	if tabWidth > 0 {
		content = FixTabs(content, tabWidth)
	} else if spacesToTab > 0 {
		content = FixSpacesToTab(content, spacesToTab)
	}
	if mode != LineEndAuto {
		content = normalizeLineEndings(content, mode)
	}
	out := FixPF2(content)
	out = FixPF1(out, targetLineEnding(mode))
	return out
}

func CheckFile(path string, mode LineEndingMode, tabWidth int, spacesToTab int) ([]Issue, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Check(path, content, mode, tabWidth, spacesToTab), nil
}

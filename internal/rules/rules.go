package rules

import "os"

const (
	PF4ID  = "PF4"
	PF4Msg = "line endings must be LF (linux-style)"
	PF5ID  = "PF5"
	PF5Msg = "line endings must be CRLF (windows-style)"
)

func Check(file string, content []byte, mode LineEndingMode) []Issue {
	var issues []Issue
	issues = append(issues, CheckPF2(file, content)...)
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

func Fix(content []byte, mode LineEndingMode) []byte {
	if mode != LineEndAuto {
		content = normalizeLineEndings(content, mode)
	}
	out := FixPF2(content)
	out = FixPF1(out, targetLineEnding(mode))
	return out
}

func CheckFile(path string, mode LineEndingMode) ([]Issue, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Check(path, content, mode), nil
}

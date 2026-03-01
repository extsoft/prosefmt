package rules

const (
	PF1ID    = "PF1"
	PF1NoEnd = "file must end with exactly one newline"
	PF1Multi = "file must end with exactly one newline (multiple newlines at end)"
)

func CheckPF1(file string, content []byte) []Issue {
	var issues []Issue
	if len(content) == 0 {
		return issues
	}
	lastLine := len(splitLines(content))
	le := detectLineEnding(content)
	if le == lineEndCRLF {
		if len(content) < 1 {
			return issues
		}
		end := len(content)
		if content[end-1] == '\n' {
			if end < 2 || content[end-2] != '\r' {
				issues = append(issues, Issue{File: file, Line: lastLine, Column: 1, RuleID: PF1ID, Message: PF1NoEnd})
				return issues
			}
			i := end - 2
			for i >= 2 && content[i-2] == '\r' && content[i-1] == '\n' {
				i -= 2
			}
			if i != end-2 {
				issues = append(issues, Issue{File: file, Line: lastLine, Column: 1, RuleID: PF1ID, Message: PF1Multi})
			}
			return issues
		}
		if content[end-1] != '\r' {
			issues = append(issues, Issue{File: file, Line: lastLine, Column: 1, RuleID: PF1ID, Message: PF1NoEnd})
			return issues
		}
		i := end - 1
		for i > 0 && content[i-1] == '\r' {
			i--
		}
		if i != end-1 {
			issues = append(issues, Issue{File: file, Line: lastLine, Column: 1, RuleID: PF1ID, Message: PF1Multi})
		}
		return issues
	}
	if content[len(content)-1] != '\n' {
		issues = append(issues, Issue{File: file, Line: lastLine, Column: 1, RuleID: PF1ID, Message: PF1NoEnd})
		return issues
	}
	i := len(content) - 1
	for i > 0 && content[i-1] == '\n' {
		i--
	}
	if i != len(content)-1 {
		issues = append(issues, Issue{File: file, Line: lastLine, Column: 1, RuleID: PF1ID, Message: PF1Multi})
	}
	return issues
}

func FixPF1(content []byte, lineEnding string) []byte {
	if len(content) == 0 {
		return content
	}
	le := lineEnding
	if le == "" {
		le = detectLineEnding(content)
	}
	if le == lineEndCRLF {
		end := len(content)
		for end >= 2 && content[end-2] == '\r' && content[end-1] == '\n' {
			end -= 2
		}
		for end > 0 && content[end-1] == '\r' {
			end--
		}
		return append(content[:end], []byte(lineEndCRLF)...)
	}
	end := len(content)
	for end > 0 && content[end-1] == '\n' {
		end--
	}
	return append(content[:end], '\n')
}

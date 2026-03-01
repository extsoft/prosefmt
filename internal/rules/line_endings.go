package rules

import "fmt"

const (
	lineEndLF   = "\n"
	lineEndCRLF = "\r\n"
)

type LineEndingMode int

const (
	LineEndAuto LineEndingMode = iota
	LineEndLinux
	LineEndWindows
)

func ParseLineEndingMode(s string) (LineEndingMode, error) {
	switch s {
	case "auto":
		return LineEndAuto, nil
	case "linux":
		return LineEndLinux, nil
	case "windows":
		return LineEndWindows, nil
	default:
		return LineEndAuto, fmt.Errorf("line endings must be auto, linux, or windows: %q", s)
	}
}

func (m LineEndingMode) String() string {
	switch m {
	case LineEndAuto:
		return "auto"
	case LineEndLinux:
		return "linux"
	case LineEndWindows:
		return "windows"
	default:
		return "auto"
	}
}

func targetLineEnding(m LineEndingMode) string {
	switch m {
	case LineEndLinux:
		return lineEndLF
	case LineEndWindows:
		return lineEndCRLF
	default:
		return ""
	}
}

func detectLineEnding(content []byte) string {
	for i := 0; i < len(content); i++ {
		if content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
			return lineEndCRLF
		}
		if content[i] == '\r' {
			return lineEndCRLF
		}
		if content[i] == '\n' {
			return lineEndLF
		}
	}
	return lineEndLF
}

func normalizeLineEndings(content []byte, mode LineEndingMode) []byte {
	if mode == LineEndAuto {
		return content
	}
	target := targetLineEnding(mode)
	lines := splitLines(content)
	var out []byte
	for i, raw := range lines {
		contentPart, _ := stripLineEnding(raw)
		out = append(out, contentPart...)
		if i < len(lines)-1 {
			out = append(out, target...)
		}
	}
	return out
}

func lineEndingMismatchLines(content []byte, mode LineEndingMode) []int {
	if mode == LineEndAuto {
		return nil
	}
	lines := splitLines(content)
	var out []int
	wantCRLF := mode == LineEndWindows
	for i, raw := range lines {
		lineNum := i + 1
		if len(raw) == 0 {
			continue
		}
		_, ending := stripLineEnding(raw)
		if wantCRLF {
			if len(ending) == 1 && ending[0] == '\n' {
				out = append(out, lineNum)
			}
		} else {
			if len(ending) == 2 && ending[0] == '\r' && ending[1] == '\n' {
				out = append(out, lineNum)
			}
			if len(ending) == 1 && ending[0] == '\r' {
				out = append(out, lineNum)
			}
		}
	}
	return out
}

func splitLines(content []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i <= len(content); i++ {
		if i == len(content) {
			lines = append(lines, content[start:i])
			break
		}
		if content[i] == '\n' {
			lines = append(lines, content[start:i+1])
			start = i + 1
		} else if content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
			lines = append(lines, content[start:i+2])
			start = i + 2
			i++
		} else if content[i] == '\r' {
			lines = append(lines, content[start:i+1])
			start = i + 1
		}
	}
	return lines
}

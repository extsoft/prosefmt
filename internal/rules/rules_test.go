package rules

import (
	"bytes"
	"testing"
)

func TestCheckPF1_NoNewline(t *testing.T) {
	content := []byte("hello")
	issues := CheckPF1("f", content)
	if len(issues) != 1 || issues[0].RuleID != PF1ID || issues[0].Message != PF1NoEnd {
		t.Errorf("expected one PF1 no-end issue, got %v", issues)
	}
}

func TestCheckPF1_OneNewlineLF(t *testing.T) {
	content := []byte("hello\n")
	issues := CheckPF1("f", content)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestCheckPF1_OneNewlineCRLF(t *testing.T) {
	content := []byte("hello\r\n")
	issues := CheckPF1("f", content)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestCheckPF1_MultipleNewlinesLF(t *testing.T) {
	content := []byte("hello\n\n")
	issues := CheckPF1("f", content)
	if len(issues) != 1 || issues[0].Message != PF1Multi {
		t.Errorf("expected one PF1 multi issue, got %v", issues)
	}
}

func TestCheckPF1_MultipleNewlinesCRLF(t *testing.T) {
	content := []byte("a\r\n\r\n")
	issues := CheckPF1("f", content)
	if len(issues) != 1 || issues[0].Message != PF1Multi {
		t.Errorf("expected one PF1 multi issue, got %v", issues)
	}
}

func TestCheckPF1_EmptyFile(t *testing.T) {
	issues := CheckPF1("f", nil)
	if len(issues) != 0 {
		t.Errorf("expected no issues for empty file, got %v", issues)
	}
}

func TestCheckPF1_EmptyFileSlice(t *testing.T) {
	issues := CheckPF1("f", []byte{})
	if len(issues) != 0 {
		t.Errorf("expected no issues for empty file slice, got %v", issues)
	}
}

func TestCheckPF1_OnlyNewline(t *testing.T) {
	content := []byte("\n")
	issues := CheckPF1("f", content)
	if len(issues) != 0 {
		t.Errorf("expected no issues for file with only newline, got %v", issues)
	}
}

func TestCheckPF1_OnlyNewlineCRLF(t *testing.T) {
	content := []byte("\r\n")
	issues := CheckPF1("f", content)
	if len(issues) != 0 {
		t.Errorf("expected no issues for file with only CRLF newline, got %v", issues)
	}
}

func TestFixPF1_NoNewline(t *testing.T) {
	content := []byte("hello")
	out := FixPF1(content, "")
	if !bytes.HasSuffix(out, []byte("\n")) || len(out) != 6 {
		t.Errorf("expected hello\\n, got %q", out)
	}
}

func TestFixPF1_OneNewline(t *testing.T) {
	content := []byte("hello\n")
	out := FixPF1(content, "")
	if !bytes.Equal(out, content) {
		t.Errorf("expected unchanged, got %q", out)
	}
}

func TestFixPF1_MultipleNewlines(t *testing.T) {
	content := []byte("hello\n\n\n")
	out := FixPF1(content, "")
	if !bytes.Equal(out, []byte("hello\n")) {
		t.Errorf("expected hello\\n, got %q", out)
	}
}

func TestFixPF1_CRLF(t *testing.T) {
	content := []byte("a\r\n\r\n")
	out := FixPF1(content, "")
	if !bytes.Equal(out, []byte("a\r\n")) {
		t.Errorf("expected a\\r\\n, got %q", out)
	}
}

func TestFixPF1_EmptyFile(t *testing.T) {
	content := []byte{}
	out := FixPF1(content, "")
	if !bytes.Equal(out, content) {
		t.Errorf("expected unchanged for empty file, got %q", out)
	}
}

func TestFixPF1_OnlyNewline(t *testing.T) {
	content := []byte("\n")
	out := FixPF1(content, "")
	if !bytes.Equal(out, content) {
		t.Errorf("expected unchanged, got %q", out)
	}
}

func TestFixPF1_OnlyNewlineCRLF(t *testing.T) {
	content := []byte("\r\n")
	out := FixPF1(content, "")
	if !bytes.Equal(out, content) {
		t.Errorf("expected unchanged, got %q", out)
	}
}

func TestCheckPF2_NoTrailing(t *testing.T) {
	content := []byte("hello\n")
	issues := CheckPF2("f", content)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestCheckPF2_TrailingSpace(t *testing.T) {
	content := []byte("hello   \n")
	issues := CheckPF2("f", content)
	if len(issues) != 1 || issues[0].Line != 1 || issues[0].Column != 6 || issues[0].RuleID != PF2ID {
		t.Errorf("expected one PF2 at line 1 col 6, got %v", issues)
	}
}

func TestCheckPF2_TrailingTab(t *testing.T) {
	content := []byte("x\t\n")
	issues := CheckPF2("f", content)
	if len(issues) != 1 || issues[0].Column != 2 {
		t.Errorf("expected one PF2 col 2, got %v", issues)
	}
}

func TestCheckPF2_MultipleLines(t *testing.T) {
	content := []byte("a\nb  \nc\n")
	issues := CheckPF2("f", content)
	if len(issues) != 1 || issues[0].Line != 2 || issues[0].Column != 2 {
		t.Errorf("expected one PF2 at line 2 col 2, got %v", issues)
	}
}

func TestCheckPF2_CRLF(t *testing.T) {
	content := []byte("hi  \r\n")
	issues := CheckPF2("f", content)
	if len(issues) != 1 || issues[0].Column != 3 {
		t.Errorf("expected one PF2 col 3, got %v", issues)
	}
}

func TestFixPF2(t *testing.T) {
	content := []byte("a  \nb\t\t\n")
	out := FixPF2(content)
	expected := []byte("a\nb\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestFixPF2_PreservesCRLF(t *testing.T) {
	content := []byte("x  \r\n")
	out := FixPF2(content)
	if !bytes.Equal(out, []byte("x\r\n")) {
		t.Errorf("expected x\\r\\n, got %q", out)
	}
}

func TestCheck_Combined(t *testing.T) {
	content := []byte("a  \nb")
	issues := Check("f", content, LineEndAuto)
	if len(issues) < 2 {
		t.Errorf("expected at least PF2 and PF1, got %v", issues)
	}
}

func TestFix_Combined(t *testing.T) {
	content := []byte("x  \n\n\n")
	out := Fix(content, LineEndAuto)
	expected := []byte("x\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
	issues := Check("f", out, LineEndAuto)
	if len(issues) != 0 {
		t.Errorf("fixed content should have no issues, got %v", issues)
	}
}

func TestCheck_LineEndingsLinux_CRLF_ReportsPF4(t *testing.T) {
	content := []byte("a\r\n")
	issues := Check("f", content, LineEndLinux)
	var pf4 []Issue
	for _, i := range issues {
		if i.RuleID == PF4ID {
			pf4 = append(pf4, i)
		}
	}
	if len(pf4) != 1 {
		t.Errorf("expected one PF4 for CRLF with linux mode, got %v", issues)
	}
}

func TestCheck_LineEndingsWindows_LF_ReportsPF5(t *testing.T) {
	content := []byte("a\n")
	issues := Check("f", content, LineEndWindows)
	var pf5 []Issue
	for _, i := range issues {
		if i.RuleID == PF5ID {
			pf5 = append(pf5, i)
		}
	}
	if len(pf5) != 1 {
		t.Errorf("expected one PF5 for LF with windows mode, got %v", issues)
	}
}

func TestCheck_LineEndings_PerLine(t *testing.T) {
	content := []byte("a\nb\r\nc\r\n")
	issues := Check("f", content, LineEndLinux)
	var pf4 []Issue
	for _, i := range issues {
		if i.RuleID == PF4ID {
			pf4 = append(pf4, i)
		}
	}
	if len(pf4) != 2 {
		t.Errorf("expected two PF4 (lines 2 and 3 have CRLF), got %d: %v", len(pf4), pf4)
	}
	lines := make(map[int]bool)
	for _, i := range pf4 {
		lines[i.Line] = true
	}
	if !lines[2] || !lines[3] {
		t.Errorf("expected PF4 at lines 2 and 3, got lines %v", lines)
	}
}

func TestFix_LineEndingsLinux_NormalizesToLF(t *testing.T) {
	content := []byte("a\r\nb  \r\n")
	out := Fix(content, LineEndLinux)
	expected := []byte("a\nb\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestFix_LineEndingsWindows_NormalizesToCRLF(t *testing.T) {
	content := []byte("a\nb  \n")
	out := Fix(content, LineEndWindows)
	expected := []byte("a\r\nb\r\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestCheckPF1_OnlyCR_ValidAsWindows(t *testing.T) {
	content := []byte("a\r")
	issues := CheckPF1("f", content)
	if len(issues) != 0 {
		t.Errorf("lone \\r (^M) is valid Windows ending, expected no issues, got %v", issues)
	}
}

func TestFixPF1_OnlyCR_NormalizesToCRLF(t *testing.T) {
	content := []byte("a\r")
	out := FixPF1(content, "")
	if !bytes.Equal(out, []byte("a\r\n")) {
		t.Errorf("expected a\\r\\n, got %q", out)
	}
}

func TestFix_LineEndingsWindows_NormalizesLoneCR(t *testing.T) {
	content := []byte("a\rb\r")
	out := Fix(content, LineEndWindows)
	expected := []byte("a\r\nb\r\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

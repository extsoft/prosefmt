package rules

import (
	"bytes"
	"testing"
)

func TestCheckPF6_NoTabs(t *testing.T) {
	content := []byte("a\nb\n")
	issues := CheckPF6("f", content, 4)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestCheckPF6_NoTabs_TabWidthZero(t *testing.T) {
	content := []byte("a\tb\n")
	issues := CheckPF6("f", content, 0)
	if len(issues) != 0 {
		t.Errorf("expected no issues when tabWidth 0, got %v", issues)
	}
}

func TestCheckPF6_FirstTabColumn(t *testing.T) {
	content := []byte("a\tb\n")
	issues := CheckPF6("f", content, 4)
	if len(issues) != 1 || issues[0].Line != 1 || issues[0].Column != 2 || issues[0].RuleID != PF6ID {
		t.Errorf("expected one PF6 line 1 col 2, got %v", issues)
	}
	if issues[0].Message != PF6Msg(4) {
		t.Errorf("message %q", issues[0].Message)
	}
}

func TestCheckPF6_OneIssuePerLine(t *testing.T) {
	content := []byte("\t\t\nx\n")
	issues := CheckPF6("f", content, 2)
	if len(issues) != 1 || issues[0].Line != 1 || issues[0].Column != 1 {
		t.Errorf("expected one issue first line col 1, got %v", issues)
	}
}

func TestCheck_PF6InCheck(t *testing.T) {
	content := []byte("hi\t\n")
	issues := Check("f", content, LineEndAuto, 4, 0)
	var pf6 int
	for _, i := range issues {
		if i.RuleID == PF6ID {
			pf6++
		}
	}
	if pf6 != 1 {
		t.Errorf("expected one PF6, issues=%v", issues)
	}
}

func TestCheck_NoPF6WhenTabWidthZero(t *testing.T) {
	content := []byte("hi\t\n")
	issues := Check("f", content, LineEndAuto, 0, 0)
	for _, i := range issues {
		if i.RuleID == PF6ID {
			t.Errorf("unexpected PF6: %v", issues)
		}
	}
}

func TestFixTabs(t *testing.T) {
	content := []byte("a\tb\tc\n")
	out := FixTabs(content, 4)
	expected := []byte("a    b    c\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestFix_ReplacesTabsBeforeOtherFixes(t *testing.T) {
	content := []byte("x\t  \n")
	out := Fix(content, LineEndAuto, 2, 0)
	expected := []byte("x\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestCheckPF7_FirstRunColumn(t *testing.T) {
	content := []byte("a    b\n")
	issues := CheckPF7("f", content, 4)
	if len(issues) != 1 || issues[0].Line != 1 || issues[0].Column != 2 || issues[0].RuleID != PF7ID {
		t.Errorf("expected one PF7 line 1 col 2, got %v", issues)
	}
}

func TestCheck_PF7InCheck(t *testing.T) {
	content := []byte("x    y\n")
	issues := Check("f", content, LineEndAuto, 0, 4)
	var pf7 int
	for _, i := range issues {
		if i.RuleID == PF7ID {
			pf7++
		}
	}
	if pf7 != 1 {
		t.Errorf("expected one PF7, issues=%v", issues)
	}
}

func TestFixSpacesToTab(t *testing.T) {
	content := []byte("a    b\n")
	out := FixSpacesToTab(content, 4)
	expected := []byte("a\tb\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestFixSpacesToTab_RepeatedUntilStable(t *testing.T) {
	content := []byte("        \n")
	out := FixSpacesToTab(content, 4)
	expected := []byte("\t\t\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestFix_SpacesToTabBeforeOtherFixes(t *testing.T) {
	content := []byte("x    \n")
	out := Fix(content, LineEndAuto, 0, 4)
	expected := []byte("x\n")
	if !bytes.Equal(out, expected) {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

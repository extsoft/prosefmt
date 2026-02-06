package report

import (
	"bytes"
	"prosefmt/internal/rules"
	"strings"
	"testing"
)

func TestWrite_Compact(t *testing.T) {
	issues := []rules.Issue{
		{File: "a.txt", Line: 1, Column: 5, RuleID: "TL010", Message: "no trailing spaces"},
		{File: "a.txt", Line: 2, Column: 1, RuleID: "TL001", Message: "file must end with exactly one newline"},
	}
	var buf bytes.Buffer
	if err := Write(&buf, FormatCompact, issues, 10, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "a.txt:1:5: TL010: no trailing spaces") {
		t.Errorf("expected compact line for first issue, got %q", out)
	}
	if !strings.Contains(out, "10 file(s) scanned, 2 issue(s).") {
		t.Errorf("expected summary with scanned count, got %q", out)
	}
}

func TestWrite_Compact_TwoFiles(t *testing.T) {
	issues := []rules.Issue{
		{File: "a.txt", Line: 1, Column: 1, RuleID: "TL001", Message: "x"},
		{File: "b.txt", Line: 1, Column: 1, RuleID: "TL001", Message: "y"},
	}
	var buf bytes.Buffer
	if err := Write(&buf, FormatCompact, issues, 6, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "6 file(s) scanned, 2 issue(s).") {
		t.Errorf("expected 6 file(s) scanned, 2 issue(s). in summary, got %q", buf.String())
	}
}

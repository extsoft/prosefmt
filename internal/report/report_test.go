package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/extsoft/prosefmt/internal/rules"
)

func TestWrite_Compact(t *testing.T) {
	issues := []rules.Issue{
		{File: "a.txt", Line: 1, Column: 5, RuleID: "PF2", Message: "no trailing spaces"},
		{File: "a.txt", Line: 2, Column: 1, RuleID: "PF1", Message: "file must end with exactly one newline"},
	}
	var buf bytes.Buffer
	if err := Write(&buf, FormatCompact, issues, 10, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "a.txt:1:5: PF2: no trailing spaces") {
		t.Errorf("expected compact line for first issue, got %q", out)
	}
	if !strings.Contains(out, "10 file(s) scanned, 2 issue(s).") {
		t.Errorf("expected summary with scanned count, got %q", out)
	}
}

func TestWrite_Compact_TwoFiles(t *testing.T) {
	issues := []rules.Issue{
		{File: "a.txt", Line: 1, Column: 1, RuleID: "PF1", Message: "x"},
		{File: "b.txt", Line: 1, Column: 1, RuleID: "PF1", Message: "y"},
	}
	var buf bytes.Buffer
	if err := Write(&buf, FormatCompact, issues, 6, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "6 file(s) scanned, 2 issue(s).") {
		t.Errorf("expected 6 file(s) scanned, 2 issue(s). in summary, got %q", buf.String())
	}
}

func TestWriteSplit_IssuesAndSummarySeparate(t *testing.T) {
	issues := []rules.Issue{
		{File: "a.txt", Line: 1, Column: 5, RuleID: "PF2", Message: "no trailing spaces"},
	}
	var issuesBuf, summaryBuf bytes.Buffer
	if err := WriteSplit(&issuesBuf, &summaryBuf, FormatCompact, issues, 3, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(issuesBuf.String(), "PF2") {
		t.Errorf("expected issues buffer to contain PF2, got %q", issuesBuf.String())
	}
	if strings.Contains(summaryBuf.String(), "PF2") {
		t.Errorf("did not expect PF2 in summary buffer, got %q", summaryBuf.String())
	}
	if !strings.Contains(summaryBuf.String(), "3 file(s) scanned, 1 issue(s).") {
		t.Errorf("expected summary in summary buffer, got %q", summaryBuf.String())
	}
}

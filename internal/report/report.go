package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/extsoft/prosefmt/internal/rules"
)

type Format string

const FormatCompact Format = "compact"

func Write(w io.Writer, format Format, issues []rules.Issue, filesScanned int, files []string) error {
	return writeCompactSplit(w, w, issues, filesScanned)
}

func WriteSplit(issuesW, summaryW io.Writer, format Format, issues []rules.Issue, filesScanned int, files []string) error {
	return writeCompactSplit(issuesW, summaryW, issues, filesScanned)
}

func writeCompactSplit(issuesW, summaryW io.Writer, issues []rules.Issue, filesScanned int) error {
	sort.Slice(issues, func(a, b int) bool {
		if issues[a].File != issues[b].File {
			return issues[a].File < issues[b].File
		}
		if issues[a].RuleID != issues[b].RuleID {
			return issues[a].RuleID < issues[b].RuleID
		}
		if issues[a].Line != issues[b].Line {
			return issues[a].Line < issues[b].Line
		}
		return issues[a].Column < issues[b].Column
	})
	for _, i := range issues {
		_, err := fmt.Fprintf(issuesW, "%s:%d:%d: %s: %s\n", i.File, i.Line, i.Column, i.RuleID, i.Message)
		if err != nil {
			return err
		}
	}
	if filesScanned >= 0 {
		_, err := fmt.Fprintf(summaryW, "%d file(s) scanned, %d issue(s).\n", filesScanned, len(issues))
		return err
	}
	files := fileSet(issues)
	_, err := fmt.Fprintf(summaryW, "%d file(s), %d issue(s).\n", len(files), len(issues))
	return err
}

func fileSet(issues []rules.Issue) map[string]bool {
	m := make(map[string]bool)
	for _, i := range issues {
		m[i.File] = true
	}
	return m
}

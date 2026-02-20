package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan_SingleFile(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	files, skipped, err := Scan([]string{good})
	if err != nil {
		t.Fatal(err)
	}
	_ = skipped
	if len(files) != 1 || files[0] != good {
		t.Errorf("expected one file %q, got %v", good, files)
	}
}

func TestScan_ExcludesBinary(t *testing.T) {
	dir := t.TempDir()
	withNull := filepath.Join(dir, "null.bin")
	if err := os.WriteFile(withNull, []byte("hello\x00world"), 0644); err != nil {
		t.Fatal(err)
	}
	files, skipped, err := Scan([]string{withNull})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files (null byte), got %v", files)
	}
	if reason, ok := skipped[withNull]; !ok || reason != "null byte" {
		t.Errorf("expected skipped[%q]=null byte, got %q", withNull, reason)
	}
}

func TestScan_ExcludesBinaryOrControlChars(t *testing.T) {
	dir := t.TempDir()
	withControl := filepath.Join(dir, "control.txt")
	if err := os.WriteFile(withControl, []byte("hello\x01world\n"), 0644); err != nil {
		t.Fatal(err)
	}
	files, skipped, err := Scan([]string{withControl})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files (control char), got %v", files)
	}
	if reason, ok := skipped[withControl]; !ok || reason != "binary or control characters" {
		t.Errorf("expected skipped[%q]=binary or control characters, got %q", withControl, reason)
	}
}

func TestScan_IncludesAnyEncoding(t *testing.T) {
	dir := t.TempDir()
	withHighBytes := filepath.Join(dir, "high.txt")
	content := []byte("r\xe9sum\xe9\n")
	if err := os.WriteFile(withHighBytes, content, 0644); err != nil {
		t.Fatal(err)
	}
	files, skipped, err := Scan([]string{withHighBytes})
	if err != nil {
		t.Fatal(err)
	}
	_ = skipped
	if len(files) != 1 || files[0] != withHighBytes {
		t.Errorf("expected one file, got %v", files)
	}
}

func TestScan_IncludesEmptyFile(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(empty, nil, 0644); err != nil {
		t.Fatal(err)
	}
	files, skipped, err := Scan([]string{empty})
	if err != nil {
		t.Fatal(err)
	}
	_ = skipped
	if len(files) != 1 || files[0] != empty {
		t.Errorf("expected one file (empty is valid text), got %v", files)
	}
}

func TestScan_RecursiveDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(sub, "b.txt")
	if err := os.WriteFile(a, []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	files, skipped, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	_ = skipped
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %v", files)
	}
}

package fix

import (
	"github.com/extsoft/prosefmt/internal/rules"
	"os"
	"path/filepath"
	"testing"
)

func TestApply(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.txt")
	content := []byte("hello   \nworld\t\t\n\n\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(path, rules.LineEndAuto); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte("hello\nworld\n")
	if string(after) != string(expected) {
		t.Errorf("expected %q, got %q", expected, after)
	}
	issues, err := rules.CheckFile(path, rules.LineEndAuto)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Errorf("fixed file should have no issues, got %v", issues)
	}
}

func TestApply_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perm.txt")
	content := []byte("hello   \n")
	if err := os.WriteFile(path, content, 0755); err != nil {
		t.Fatal(err)
	}
	infoBefore, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(path, rules.LineEndAuto); err != nil {
		t.Fatal(err)
	}
	infoAfter, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if infoAfter.Mode().Perm() != infoBefore.Mode().Perm() {
		t.Errorf("expected permissions %v, got %v", infoBefore.Mode().Perm(), infoAfter.Mode().Perm())
	}
}

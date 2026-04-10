package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_Check_TabsToSpaces_PF6(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a\tz\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", "--replace-tabs-with-spaces", "4", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("expected exit 1, got %d\n%s", cmd.ProcessState.ExitCode(), out)
	}
	if !strings.Contains(string(out), "PF6") {
		t.Errorf("expected PF6 in output, got %s", out)
	}
}

func TestIntegration_Check_NoPF6WithoutFlag(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a\tz\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("expected exit 0, got %d\n%s", cmd.ProcessState.ExitCode(), out)
	}
	if strings.Contains(string(out), "PF6") {
		t.Errorf("did not expect PF6, got %s", out)
	}
}

func TestIntegration_Write_TabsToSpaces(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a\tb\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "write", "--replace-tabs-with-spaces", "2", p)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("write: %v", err)
	}
	after, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "a  b\n" {
		t.Errorf("expected \"a  b\\n\", got %q", after)
	}
}

func TestIntegration_TabsToSpaces_InvalidZero(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", "--replace-tabs-with-spaces", "0", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() == 0 {
		t.Errorf("expected non-zero exit for replace-tabs-with-spaces 0, got %s", out)
	}
}

func TestIntegration_TabsToSpaces_InvalidNegative(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", "--replace-tabs-with-spaces", "-1", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() == 0 {
		t.Errorf("expected non-zero exit for replace-tabs-with-spaces -1, got %s", out)
	}
}

func TestIntegration_Check_SpacesToTab_PF7(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a    z\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", "--replace-spaces-with-tabs", "4", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("expected exit 1, got %d\n%s", cmd.ProcessState.ExitCode(), out)
	}
	if !strings.Contains(string(out), "PF7") {
		t.Errorf("expected PF7 in output, got %s", out)
	}
}

func TestIntegration_Check_NoPF7WithoutFlag(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a    z\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("expected exit 0, got %d\n%s", cmd.ProcessState.ExitCode(), out)
	}
	if strings.Contains(string(out), "PF7") {
		t.Errorf("did not expect PF7, got %s", out)
	}
}

func TestIntegration_Write_SpacesToTab(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a    b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "write", "--replace-spaces-with-tabs", "4", p)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("write: %v", err)
	}
	after, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "a\tb\n" {
		t.Errorf("expected \"a\\tb\\n\", got %q", after)
	}
}

func TestIntegration_TabFlags_MutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.txt")
	if err := os.WriteFile(p, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exe := buildBinary(t)
	cmd := exec.Command(exe, "check", "--replace-tabs-with-spaces", "4", "--replace-spaces-with-tabs", "4", p)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() == 0 {
		t.Errorf("expected non-zero exit when both tab flags set, got %s", out)
	}
}

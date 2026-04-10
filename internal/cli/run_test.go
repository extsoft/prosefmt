package cli_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/extsoft/prosefmt/internal/cli"
	"github.com/extsoft/prosefmt/internal/log"
	"github.com/extsoft/prosefmt/internal/rules"
)

func captureStdout(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	old := os.Stdout
	os.Stdout = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		io.Copy(&buf, r)
		r.Close()
		close(done)
	}()
	fn()
	w.Close()
	os.Stdout = old
	<-done
	return buf.String()
}

func captureStdoutStderr(fn func()) (stdout, stderr string) {
	rOut, wOut, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr
	log.SetOutput(wOut)
	doneOut := make(chan struct{})
	doneErr := make(chan struct{})
	var bufOut, bufErr bytes.Buffer
	go func() {
		io.Copy(&bufOut, rOut)
		rOut.Close()
		close(doneOut)
	}()
	go func() {
		io.Copy(&bufErr, rErr)
		rErr.Close()
		close(doneErr)
	}()
	fn()
	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	log.SetOutput(nil)
	<-doneOut
	<-doneErr
	return bufOut.String(), bufErr.String()
}

func TestRun_Silent_NoStdout(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(f, []byte("x  \n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Silent)
	defer log.SetLevel(log.Normal)
	var hadIssues bool
	var runErr error
	stdout, stderr := captureStdoutStderr(func() {
		hadIssues, runErr = cli.Run(true, false, []string{f}, rules.LineEndAuto, 0, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !hadIssues {
		t.Error("expected hadIssues true")
	}
	if len(stdout) != 0 || len(stderr) != 0 {
		t.Errorf("silent: expected no stdout/stderr output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestRun_Normal_StderrHasIssuesAndSummary(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(f, []byte("x  \n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Normal)
	defer log.SetLevel(log.Normal)
	var runErr error
	stdout, stderr := captureStdoutStderr(func() {
		_, runErr = cli.Run(true, false, []string{f}, rules.LineEndAuto, 0, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if len(stdout) != 0 {
		t.Errorf("normal: expected no stdout when issues present, got %q", stdout)
	}
	if !strings.Contains(stderr, "file(s)") || !strings.Contains(stderr, "issue(s)") {
		t.Errorf("normal: expected report summary on stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "PF2") {
		t.Errorf("normal: expected rule ID on stderr, got %q", stderr)
	}
}

func TestRun_Verbose_StdoutHasScanning(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(f, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Verbose)
	defer log.SetLevel(log.Normal)
	var runErr error
	stdout, stderr := captureStdoutStderr(func() {
		_, runErr = cli.Run(true, false, []string{f}, rules.LineEndAuto, 0, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !strings.Contains(stdout, "Scanning") && !strings.Contains(stdout, "Scanned") {
		t.Errorf("verbose: expected Scanning/Scanned on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "Completed in") {
		t.Errorf("verbose: expected timing on stdout, got %q", stdout)
	}
	if len(stderr) != 0 {
		t.Errorf("verbose: expected no stderr without PF issues, got %q", stderr)
	}
}

func TestRun_Verbose_StdoutHasRejectedAndAccepted(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	bin := filepath.Join(dir, "bin.bin")
	if err := os.WriteFile(good, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte("x\x00y"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Verbose)
	defer log.SetLevel(log.Normal)
	var runErr error
	stdout, stderr := captureStdoutStderr(func() {
		_, runErr = cli.Run(true, false, []string{dir}, rules.LineEndAuto, 0, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !strings.Contains(stdout, "Configuration:") {
		t.Errorf("verbose: expected Configuration line on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "scanner: rejected") {
		t.Errorf("verbose: expected scanner rejected on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "scanner: accepted") {
		t.Errorf("verbose: expected scanner accepted on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "null byte") {
		t.Errorf("verbose: expected reason null byte for binary file on stdout, got %q", stdout)
	}
	if len(stderr) != 0 {
		t.Errorf("verbose: expected no stderr without PF issues, got %q", stderr)
	}
}

func TestRun_ZeroTextFiles_Normal_NoTextFilesFound(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "x.bin")
	if err := os.WriteFile(bin, []byte("\x00"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Normal)
	defer log.SetLevel(log.Normal)
	var runErr error
	stdout := captureStdout(func() {
		_, runErr = cli.Run(true, false, []string{bin}, rules.LineEndAuto, 0, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !strings.Contains(stdout, "No text files found.") {
		t.Errorf("expected No text files found. on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "0 file(s) scanned, 0 issue(s)") {
		t.Errorf("expected 0 file(s) scanned, 0 issue(s) in summary, got %q", stdout)
	}
}

func TestRun_TabWidth_ReportsPF6(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "tab.txt")
	if err := os.WriteFile(f, []byte("a\tb\n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Normal)
	defer log.SetLevel(log.Normal)
	var hadIssues bool
	var runErr error
	stdout, stderr := captureStdoutStderr(func() {
		hadIssues, runErr = cli.Run(true, false, []string{f}, rules.LineEndAuto, 4, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !hadIssues {
		t.Error("expected hadIssues true for tabs with tabWidth 4")
	}
	if !strings.Contains(stderr, "PF6") {
		t.Errorf("expected PF6 on stderr, got stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, "file(s) scanned") {
		t.Errorf("expected scan summary on stderr with issues, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestRun_TabWidthZero_NoPF6ForTabs(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "tab.txt")
	if err := os.WriteFile(f, []byte("a\tb\n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Normal)
	defer log.SetLevel(log.Normal)
	var hadIssues bool
	var runErr error
	stdout, stderr := captureStdoutStderr(func() {
		hadIssues, runErr = cli.Run(true, false, []string{f}, rules.LineEndAuto, 0, 0)
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if hadIssues {
		t.Errorf("expected no issues without replace-tabs-with-spaces, got hadIssues=true stdout=%q stderr=%q", stdout, stderr)
	}
	if strings.Contains(stdout, "PF6") || strings.Contains(stderr, "PF6") {
		t.Errorf("did not expect PF6, got stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "1 file(s) scanned, 0 issue(s)") {
		t.Errorf("expected zero-issue summary on stdout, got stdout=%q stderr=%q", stdout, stderr)
	}
	if len(stderr) != 0 {
		t.Errorf("expected no stderr without issues, got %q", stderr)
	}
}

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

func captureStderr(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	old := os.Stderr
	os.Stderr = w
	log.SetOutput(w)
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		io.Copy(&buf, r)
		r.Close()
		close(done)
	}()
	fn()
	w.Close()
	os.Stderr = old
	log.SetOutput(nil)
	<-done
	return buf.String()
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
	stdout := captureStdout(func() {
		hadIssues, runErr = cli.Run(true, false, []string{f})
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !hadIssues {
		t.Error("expected hadIssues true")
	}
	if len(stdout) != 0 {
		t.Errorf("silent: expected no stdout, got %q", stdout)
	}
}

func TestRun_Normal_StdoutHasReport(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(f, []byte("x  \n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Normal)
	defer log.SetLevel(log.Normal)
	var runErr error
	stdout := captureStdout(func() {
		_, runErr = cli.Run(true, false, []string{f})
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !strings.Contains(stdout, "file(s)") || !strings.Contains(stdout, "issue(s)") {
		t.Errorf("normal: expected report summary on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "TL010") {
		t.Errorf("normal: expected rule ID in output, got %q", stdout)
	}
}

func TestRun_Verbose_StderrHasScanning(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(f, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	log.SetLevel(log.Verbose)
	defer log.SetLevel(log.Normal)
	var runErr error
	stderr := captureStderr(func() {
		_, runErr = cli.Run(true, false, []string{f})
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !strings.Contains(stderr, "Scanning") && !strings.Contains(stderr, "Scanned") {
		t.Errorf("verbose: expected Scanning/Scanned on stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "Completed in") {
		t.Errorf("verbose: expected timing on stderr, got %q", stderr)
	}
}

func TestRun_Verbose_StderrHasRejectedAndAccepted(t *testing.T) {
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
	stderr := captureStderr(func() {
		_, runErr = cli.Run(true, false, []string{dir})
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !strings.Contains(stderr, "Configuration:") {
		t.Errorf("verbose: expected Configuration line on stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "scanner: rejected") {
		t.Errorf("verbose: expected scanner rejected on stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "scanner: accepted") {
		t.Errorf("verbose: expected scanner accepted on stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "null byte") {
		t.Errorf("verbose: expected reason null byte for binary file, got %q", stderr)
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
		_, runErr = cli.Run(true, false, []string{bin})
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

package log

import (
	"bytes"
	"testing"
)

func TestLogf_RespectsLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	defer SetOutput(nil)
	SetLevel(Silent)
	Logf(Verbose, "should not appear\n")
	if buf.Len() != 0 {
		t.Errorf("Silent: expected no output for Verbose, got %q", buf.String())
	}
	buf.Reset()
	SetLevel(Verbose)
	Logf(Verbose, "visible\n")
	if !bytes.Contains(buf.Bytes(), []byte("visible")) {
		t.Errorf("Verbose: expected visible, got %q", buf.String())
	}
	SetLevel(Normal)
}

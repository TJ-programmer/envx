package cli

import (
	"strings"
	"testing"
)

func TestWebRequiresInitializedProject(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := runIn(dir, "web")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "not initialized") {
		t.Fatalf("stderr should mention not initialized: %s", errOut)
	}
}

func TestWebRejectsInvalidPort(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	code, _, errOut := runIn(dir, "web", "--port", "notaport")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "invalid port") {
		t.Fatalf("stderr should mention invalid port: %s", errOut)
	}

	code, _, errOut = runIn(dir, "web", "--port", "99999")
	if code != 1 || !strings.Contains(errOut, "invalid port") {
		t.Fatalf("out-of-range port should be rejected, code=%d err=%s", code, errOut)
	}
}

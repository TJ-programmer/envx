package cli

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWebServesWithoutInitializedProject(t *testing.T) {
	_ = t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	webSignalCtx = func() context.Context { return ctx }
	defer cancel()

	var stdout, stderr strings.Builder
	codeCh := make(chan int, 1)
	go func() {
		codeCh <- Run([]string{"web", "--no-open", "--port", "46313"}, &stdout, &stderr)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:46313", 200*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never started: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	cancel()
	select {
	case code := <-codeCh:
		if code != 0 {
			t.Fatalf("code = %d, want 0", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down after cancel")
	}

	out := stdout.String()
	if !strings.Contains(out, "not initialized") {
		t.Fatalf("stdout should mention not initialized: %s", out)
	}
	if !strings.Contains(out, "46313") {
		t.Fatalf("stdout should mention the port: %s", out)
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

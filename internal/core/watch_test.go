package core_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"envx/internal/bootstrap"
)

func TestRunWatchRestartsOnChange(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER") == "1" {
		watchHelperLoop()
		return
	}
	dir := t.TempDir()
	outFile := filepath.Join(dir, "seen.txt")
	stopFile := filepath.Join(dir, "stop.txt")
	service := bootstrap.BuildService(dir)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	for _, kv := range [][2]string{
		{"WATCH_VAR", "v1"},
		{"WATCH_OUT", outFile},
		{"WATCH_STOP", stopFile},
	} {
		if _, err := service.SetVariable(kv[0], kv[1], "", false, false); err != nil {
			t.Fatal(err)
		}
	}

	argv := []string{os.Args[0], "-test.run=TestRunWatchRestartsOnChange"}
	paths := []string{
		filepath.Join(dir, ".env"),
		filepath.Join(dir, ".envx", "config.json"),
		filepath.Join(dir, ".envx", "key.bin"),
	}
	codeCh := make(chan int, 1)
	go func() {
		os.Setenv("GO_WANT_HELPER", "1")
		code, _ := service.RunWatch(argv, "", "dev", false, 20*time.Millisecond, io.Discard, paths)
		codeCh <- code
	}()

	waitForContent(t, outFile, "v1", 10*time.Second)

	if _, err := service.SetVariable("WATCH_VAR", "v2", "", false, false); err != nil {
		t.Fatal(err)
	}
	waitForContent(t, outFile, "v2", 10*time.Second)

	if err := os.WriteFile(stopFile, []byte("stop"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case code := <-codeCh:
		if code != 0 {
			t.Fatalf("watch exit code = %d, want 0", code)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("watch loop did not stop after stop file")
	}
}

func waitForContent(t *testing.T, path, needle string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(data), needle) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	data, _ := os.ReadFile(path)
	t.Fatalf("timed out waiting for %q in %s (got %q)", needle, path, string(data))
}

func watchHelperLoop() {
	out, err := os.OpenFile(os.Getenv("WATCH_OUT"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		os.Exit(1)
	}
	stop := os.Getenv("WATCH_STOP")
	for {
		fmt.Fprintln(out, os.Getenv("WATCH_VAR"))
		out.Sync()
		if stop != "" {
			if _, err := os.Stat(stop); err == nil {
				os.Exit(0)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

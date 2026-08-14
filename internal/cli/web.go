package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"

	"envx/internal/web"
)

func cmdWeb(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "web") {
		return 0
	}
	port := 4319
	openBrowser := true
	for len(rest) > 0 {
		switch rest[0] {
		case "--port":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--port'"))
			}
			parsed, err := strconv.Atoi(rest[1])
			if err != nil || parsed < 1 || parsed > 65535 {
				return printError(stderr, fmt.Errorf("invalid port '%s'", rest[1]))
			}
			port = parsed
			rest = rest[2:]
		case "--no-open":
			openBrowser = false
			rest = rest[1:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := projectService(root)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := "http://" + addr
	if !service.IsInitialized() {
		fmt.Fprintln(stdout, "Project is not initialized; you can initialize it from the web UI.")
	}
	fmt.Fprintf(stdout, "envx web running at %s\n", url)
	fmt.Fprintf(stdout, "Press Ctrl+C to stop.\n")
	if openBrowser {
		openInBrowser(url)
	}

	server := &http.Server{Addr: addr, Handler: web.New(service)}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return printError(stderr, fmt.Errorf("failed to start web server: %v", err))
		}
		return 0
	case <-webSignalCtx().Done():
		server.Close()
		return 0
	}
}

func signalContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

var webSignalCtx = signalContext

func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

package main

import (
	"os"
	"os/signal"
	"syscall"

	"envx/internal/cli"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		os.Stderr.WriteString("\n")
		os.Exit(130)
	}()

	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}

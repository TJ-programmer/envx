package buildinfo

// Version is the CLI version. Override at build time with:
//
//	go build -ldflags "-X envx/internal/buildinfo.Version=v1.2.3"
var Version = "0.5.0"

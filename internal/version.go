package internal

import (
	"fmt"
	"runtime"
)

const notProvided = "[not provided]"

var (
	// Version is the git tag this version is built with (injected at build time)
	Version = "development"
	// BuildDate is the time this version was built
	BuildDate = notProvided
	// Platform is the architecture this is running on
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	// GoVersion of Go built with
	GoVersion = runtime.Version()
)

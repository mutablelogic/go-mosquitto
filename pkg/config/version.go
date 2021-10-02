package config

import (
	"fmt"
	"io"
	"runtime"

	// Packages
	mosq "github.com/djthorpe/go-mosquitto/sys/mosquitto"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GitHash     string
	GoBuildTime string
)

func PrintVersion(w io.Writer) {
	if GitSource != "" {
		fmt.Fprintf(w, "  URL: https://%v\n", GitSource)
	}
	if GitTag != "" || GitBranch != "" {
		fmt.Fprintf(w, "  Version: %v (branch: %q hash:%q)\n", GitTag, GitBranch, GitHash)
	}
	if GoBuildTime != "" {
		fmt.Fprintf(w, "  Build Time: %v\n", GoBuildTime)
	}
	fmt.Fprintf(w, "  Go: %v (%v/%v)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(w, "  libmosquitto: %v\n", LibVersion())
}

func LibVersion() string {
	major, minor, revision := mosq.Version()
	return fmt.Sprintf("%d.%d.%d", major, minor, revision)
}

// Package version holds build-time metadata for all Archon AI binaries.
//
// The variables are intentionally empty by default and are stamped at build
// time via -ldflags, for example:
//
//	go build -ldflags "-X github.com/hacrex/Archon-AI/internal/version.Version=v0.1.0 \
//	  -X github.com/hacrex/Archon-AI/internal/version.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/hacrex/Archon-AI/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//	  ./cmd/archon
package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Values overridden at build time. Keep these three in sync with the
// LDFLAGS defined in the repository Makefile.
var (
	// Version is the semantic version or git tag of the build.
	Version = "0.0.1-dev"
	// Commit is the short SHA of the git revision the binary was built from.
	Commit = "unknown"
	// BuildDate is the UTC timestamp of the build (RFC3339).
	BuildDate = "unknown"
)

// Info summarizes the build metadata of a running binary.
type Info struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns the build info for the named binary (e.g. "archon",
// "archon-api", "archon-monitor").
func Get(name string) Info {
	return Info{
		Name:      name,
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String renders a single-line summary such as:
//
//	archon 0.1.0 (commit abc1234, built 2026-11-03T00:00:00Z, go1.22.10 linux/amd64)
func (i Info) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", i.Name, i.Version)
	fmt.Fprintf(&b, " (commit %s, built %s, %s %s)", i.Commit, i.BuildDate, i.GoVersion, i.Platform)
	return b.String()
}

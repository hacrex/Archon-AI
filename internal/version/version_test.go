package version

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetDefaults(t *testing.T) {
	info := Get("archon")
	if info.Name != "archon" {
		t.Errorf("Name = %q, want %q", info.Name, "archon")
	}
	if info.Version == "" {
		t.Error("Version must never be empty, even without ldflags")
	}
	if !strings.Contains(info.GoVersion, "go1.") {
		t.Errorf("GoVersion = %q, want a go1.x string", info.GoVersion)
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform = %q, want os/arch", info.Platform)
	}
}

func TestStringFormat(t *testing.T) {
	Version = "v9.9.9"
	Commit = "abc1234"
	BuildDate = "2026-11-03T00:00:00Z"
	t.Cleanup(func() { Version, Commit, BuildDate = "0.0.1-dev", "unknown", "unknown" })

	s := Get("archon-api").String()
	for _, want := range []string{"archon-api v9.9.9", "commit abc1234", "built 2026-11-03T00:00:00Z"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() = %q, missing %q", s, want)
		}
	}
}

func TestInfoJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(Get("archon-monitor"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Info
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Name != "archon-monitor" || back.Version != Version {
		t.Errorf("round trip mismatch: %+v", back)
	}
}

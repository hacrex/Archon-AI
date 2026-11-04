package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevStartsWaitsWritesStateAndDevDownStops(t *testing.T) {
	changeWorkingDirectory(t, t.TempDir())
	composeFile := filepath.Join("deploy", "docker-compose.yml")
	if err := os.MkdirAll(filepath.Dir(composeFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(composeFile, []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	logFile := filepath.Join(t.TempDir(), "docker-args.log")
	installFakeDocker(t, logFile)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/livez" || r.URL.Path == "/readyz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer api.Close()

	if err := dev([]string{"--compose-file", composeFile, "--api-url", api.URL, "--timeout", "2s"}); err != nil {
		t.Fatalf("dev() error = %v", err)
	}
	state, err := readDevState()
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.ComposeFile != composeFile || state.APIURL != api.URL {
		t.Fatalf("unexpected state: %#v", state)
	}
	info, err := os.Stat(filepath.Join(devStateDir, devStateFile))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("state permissions = %o, want 600", info.Mode().Perm())
	}

	if err := dev([]string{"down"}); err != nil {
		t.Fatalf("dev down error = %v", err)
	}
	if _, err := os.Stat(devStateDir); !os.IsNotExist(err) {
		t.Fatalf("state directory still exists: %v", err)
	}
	log, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	text := string(log)
	if !strings.Contains(text, "compose -f "+composeFile+" up -d --build") || !strings.Contains(text, "compose -f "+composeFile+" down") {
		t.Fatalf("unexpected docker invocations: %s", text)
	}
}

func TestDevReadinessFailureStopsServices(t *testing.T) {
	changeWorkingDirectory(t, t.TempDir())
	composeFile := "docker-compose.yml"
	if err := os.WriteFile(composeFile, []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	logFile := filepath.Join(t.TempDir(), "docker-args.log")
	installFakeDocker(t, logFile)
	api := httptest.NewServer(http.NotFoundHandler())
	defer api.Close()

	err := dev([]string{"--compose-file", composeFile, "--api-url", api.URL, "--timeout", "200ms"})
	if err == nil || !strings.Contains(err.Error(), "services were stopped") {
		t.Fatalf("dev() error = %v, want readiness cleanup error", err)
	}
	log, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(log), "down") {
		t.Fatalf("cleanup did not invoke docker down: %s", log)
	}
}

func installFakeDocker(t *testing.T, logFile string) {
	t.Helper()
	binDir := t.TempDir()
	script := filepath.Join(binDir, "docker")
	content := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> %q\nexit 0\n", logFile)
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func changeWorkingDirectory(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

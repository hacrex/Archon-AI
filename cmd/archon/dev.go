package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultComposeFile = "deploy/docker-compose.yml"
	defaultAPIURL      = "http://127.0.0.1:8080"
	devStateDir        = ".archon"
	devStateFile       = "state.json"
)

type devState struct {
	ComposeFile string    `json:"composeFile"`
	APIURL      string    `json:"apiUrl"`
	StartedAt   time.Time `json:"startedAt"`
	PID         int       `json:"pid"`
}

func dev(args []string) error {
	if len(args) > 0 && args[0] == "down" {
		return devDown(args[1:])
	}
	flags := newFlagSet("dev")
	timeout := flags.Duration("timeout", 2*time.Minute, "maximum time to wait for the local API")
	composeFile := flags.String("compose-file", envOr("ARCHON_COMPOSE_FILE", defaultComposeFile), "Docker Compose file")
	apiURL := flags.String("api-url", envOr("ARCHON_API_URL", defaultAPIURL), "local API base URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}
	if _, err := os.Stat(*composeFile); err != nil {
		return fmt.Errorf("compose file %q is unavailable: %w", *composeFile, err)
	}
	if err := ensureDevStateDir(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := runCompose(ctx, *composeFile, "up", "-d", "--build"); err != nil {
		return fmt.Errorf("start local services: %w", err)
	}
	fmt.Printf("started local services; waiting for %s\n", strings.TrimRight(*apiURL, "/"))
	if err := waitForAPI(ctx, *apiURL); err != nil {
		_ = runCompose(context.Background(), *composeFile, "down")
		return fmt.Errorf("local API did not become ready: %w; services were stopped", err)
	}
	state := devState{ComposeFile: *composeFile, APIURL: strings.TrimRight(*apiURL, "/"), StartedAt: time.Now().UTC(), PID: os.Getpid()}
	if err := writeDevState(state); err != nil {
		return err
	}
	fmt.Printf("Archon is ready at %s\n", state.APIURL)
	fmt.Println("Next: archon bootstrap-user --email owner@example.com --display-name 'Local Owner' --password-stdin")
	fmt.Println("Stop: archon dev down")
	return nil
}

func devDown(args []string) error {
	flags := newFlagSet("dev down")
	composeFile := flags.String("compose-file", envOr("ARCHON_COMPOSE_FILE", ""), "Docker Compose file")
	volumes := flags.Bool("volumes", false, "remove local Compose volumes")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	state, stateErr := readDevState()
	if *composeFile == "" && stateErr == nil {
		*composeFile = state.ComposeFile
	}
	if *composeFile == "" {
		*composeFile = defaultComposeFile
	}
	if _, err := os.Stat(*composeFile); err != nil {
		return fmt.Errorf("compose file %q is unavailable: %w", *composeFile, err)
	}
	argsForCompose := []string{"down"}
	if *volumes {
		argsForCompose = append(argsForCompose, "--volumes")
	}
	if err := runCompose(context.Background(), *composeFile, argsForCompose...); err != nil {
		return fmt.Errorf("stop local services: %w", err)
	}
	if err := removeDevState(); err != nil {
		return err
	}
	fmt.Println("Archon local services stopped")
	return nil
}

func runCompose(ctx context.Context, composeFile string, args ...string) error {
	composeArgs := append([]string{"compose", "-f", composeFile}, args...)
	command := exec.CommandContext(ctx, "docker", composeArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("docker compose timed out")
		}
		return fmt.Errorf("docker compose %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func waitForAPI(ctx context.Context, baseURL string) error {
	baseURL = strings.TrimRight(baseURL, "/")
	client := &http.Client{Timeout: 3 * time.Second}
	var lastErr error
	for {
		if err := checkEndpoint(ctx, client, baseURL+"/livez"); err != nil {
			lastErr = err
		} else if err := checkEndpoint(ctx, client, baseURL+"/readyz"); err != nil {
			lastErr = err
		} else {
			return nil
		}
		select {
		case <-ctx.Done():
			if lastErr == nil {
				lastErr = ctx.Err()
			}
			return lastErr
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func checkEndpoint(ctx context.Context, client *http.Client, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s returned HTTP %s", endpoint, response.Status)
	}
	return nil
}

func ensureDevStateDir() error {
	if err := os.MkdirAll(devStateDir, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", devStateDir, err)
	}
	return os.Chmod(devStateDir, 0o700)
}

func writeDevState(state devState) error {
	if err := ensureDevStateDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode dev state: %w", err)
	}
	path := filepath.Join(devStateDir, devStateFile)
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write dev state: %w", err)
	}
	return nil
}

func readDevState() (devState, error) {
	data, err := os.ReadFile(filepath.Join(devStateDir, devStateFile))
	if err != nil {
		return devState{}, err
	}
	var state devState
	if err := json.Unmarshal(data, &state); err != nil {
		return devState{}, fmt.Errorf("decode dev state: %w", err)
	}
	if state.ComposeFile == "" {
		return devState{}, errors.New("dev state has no compose file")
	}
	return state, nil
}

func removeDevState() error {
	path := filepath.Join(devStateDir, devStateFile)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove dev state: %w", err)
	}
	if err := os.Remove(devStateDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove dev state directory: %w", err)
	}
	return nil
}

func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

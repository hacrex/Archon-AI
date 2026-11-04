package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const (
	minimumMemoryBytes = uint64(8 * 1024 * 1024 * 1024)
	minimumDiskBytes   = uint64(20 * 1024 * 1024 * 1024)
	minimumCPUs        = 2
)

type doctorCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Required    bool   `json:"required"`
	Summary     string `json:"summary"`
	Detail      string `json:"detail,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

type doctorReport struct {
	OK      bool          `json:"ok"`
	Checks  []doctorCheck `json:"checks"`
	Version string        `json:"version"`
}

func doctor(args []string) error {
	return doctorWithIO(args, os.Stdout)
}

func doctorWithIO(args []string, out io.Writer) error {
	flags := newFlagSet("doctor")
	output := flags.String("output", "text", "output format: text or json")
	apiURL := flags.String("api-url", envOr("ARCHON_API_URL", defaultAPIURL), "API URL to check for a port conflict")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *output != "text" && *output != "json" {
		return errors.New("--output must be text or json")
	}

	report := collectDoctorReport(*apiURL)
	if *output == "json" {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("encode doctor report: %w", err)
		}
	} else {
		writeDoctorText(out, report)
	}
	if !report.OK {
		return errors.New("one or more required checks failed")
	}
	return nil
}

func collectDoctorReport(apiURL string) doctorReport {
	report := doctorReport{OK: true, Version: envOr("ARCHON_VERSION", "development")}
	report.Checks = append(report.Checks,
		checkCommand("docker", "Docker is available", true),
		checkCompose(),
		checkCommand("python3", "Python 3 is available", true),
		checkCPU(),
		checkMemory(),
		checkDisk(),
		checkPort(apiURL),
		checkWritableDirectory(),
	)
	for _, check := range report.Checks {
		if check.Required && check.Status == "fail" {
			report.OK = false
		}
	}
	return report
}

func checkCommand(name, summary string, required bool) doctorCheck {
	check := doctorCheck{Name: name, Summary: summary, Required: required}
	path, err := exec.LookPath(name)
	if err != nil {
		check.Status = "fail"
		check.Detail = "command was not found on PATH"
		check.Remediation = fmt.Sprintf("install %s and make it available on PATH", name)
		return check
	}
	check.Status = "pass"
	check.Detail = path
	return check
}

func checkCompose() doctorCheck {
	check := doctorCheck{Name: "docker-compose", Summary: "Docker Compose v2 is available", Required: true}
	if _, err := exec.LookPath("docker"); err != nil {
		check.Status = "fail"
		check.Detail = "Docker is not available, so Compose cannot be checked"
		check.Remediation = "install Docker Engine and the Compose v2 plugin"
		return check
	}
	command := exec.Command("docker", "compose", "version")
	output, err := command.CombinedOutput()
	if err != nil {
		check.Status = "fail"
		check.Detail = strings.TrimSpace(string(output))
		check.Remediation = "install or enable the Docker Compose v2 plugin"
		return check
	}
	check.Status = "pass"
	check.Detail = strings.TrimSpace(string(output))
	return check
}

func checkCPU() doctorCheck {
	count := runtime.NumCPU()
	check := doctorCheck{Name: "cpu", Summary: "Host has at least two logical CPUs", Required: false, Detail: strconv.Itoa(count) + " logical CPUs"}
	if count < minimumCPUs {
		check.Status = "warn"
		check.Remediation = "use a host with at least 2 logical CPUs for the recommended profile"
		return check
	}
	check.Status = "pass"
	return check
}

func checkMemory() doctorCheck {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return doctorCheck{Name: "memory", Status: "warn", Summary: "Host memory can be inspected", Required: false, Detail: "memory information is unavailable on this operating system", Remediation: "verify at least 8 GiB RAM manually"}
	}
	var memoryKB uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "MemTotal:" {
			memoryKB, _ = strconv.ParseUint(fields[1], 10, 64)
			break
		}
	}
	if memoryKB == 0 {
		return doctorCheck{Name: "memory", Status: "warn", Summary: "Host memory can be inspected", Required: false, Detail: "MemTotal is unavailable", Remediation: "verify at least 8 GiB RAM manually"}
	}
	bytes := memoryKB * 1024
	check := doctorCheck{Name: "memory", Summary: "Host has the recommended memory", Required: false, Detail: fmt.Sprintf("%.1f GiB", float64(bytes)/(1024*1024*1024))}
	if bytes < minimumMemoryBytes {
		check.Status = "warn"
		check.Remediation = "use a host with at least 8 GiB RAM for the recommended profile"
		return check
	}
	check.Status = "pass"
	return check
}

func checkDisk() doctorCheck {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(".", &stat); err != nil {
		return doctorCheck{Name: "disk", Status: "warn", Summary: "Writable disk space can be inspected", Required: false, Detail: err.Error(), Remediation: "verify at least 20 GiB free disk space manually"}
	}
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	check := doctorCheck{Name: "disk", Summary: "Host has the recommended free disk space", Required: false, Detail: fmt.Sprintf("%.1f GiB free", float64(free)/(1024*1024*1024))}
	if free < minimumDiskBytes {
		check.Status = "warn"
		check.Remediation = "free at least 20 GiB before pulling images and creating volumes"
		return check
	}
	check.Status = "pass"
	return check
}

func checkPort(apiURL string) doctorCheck {
	check := doctorCheck{Name: "api-port", Summary: "The configured API port is available", Required: true}
	address := strings.TrimPrefix(strings.TrimPrefix(apiURL, "http://"), "https://")
	if slash := strings.IndexByte(address, '/'); slash >= 0 {
		address = address[:slash]
	}
	if !strings.Contains(address, ":") {
		address += ":80"
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		check.Status = "fail"
		check.Detail = fmt.Sprintf("cannot listen on %s: %v", address, err)
		check.Remediation = "stop the process using this port or choose another local API port"
		return check
	}
	_ = listener.Close()
	check.Status = "pass"
	check.Detail = address
	return check
}

func checkWritableDirectory() doctorCheck {
	check := doctorCheck{Name: "workspace", Summary: "The current workspace is writable", Required: true}
	file, err := os.CreateTemp(".", ".archon-doctor-*")
	if err != nil {
		check.Status = "fail"
		check.Detail = err.Error()
		check.Remediation = "run Archon from a writable project directory"
		return check
	}
	path := file.Name()
	_ = file.Close()
	_ = os.Remove(path)
	check.Status = "pass"
	check.Detail, _ = filepath.Abs(".")
	return check
}

func writeDoctorText(out io.Writer, report doctorReport) {
	fmt.Fprintf(out, "Archon doctor (%s)\n\n", report.Version)
	for _, check := range report.Checks {
		marker := "OK"
		if check.Status == "warn" {
			marker = "WARN"
		} else if check.Status == "fail" {
			marker = "FAIL"
		}
		fmt.Fprintf(out, "[%s] %-16s %s\n", marker, check.Name, check.Summary)
		if check.Detail != "" {
			fmt.Fprintf(out, "      %s\n", check.Detail)
		}
		if check.Status != "pass" && check.Remediation != "" {
			fmt.Fprintf(out, "      Fix: %s\n", check.Remediation)
		}
	}
	if report.OK {
		fmt.Fprintln(out, "\nRequired checks passed.")
	} else {
		fmt.Fprintln(out, "\nRequired checks failed.")
	}
}

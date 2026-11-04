package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDoctorJSONReportIsMachineReadable(t *testing.T) {
	var output bytes.Buffer
	err := doctorWithIO([]string{"--output", "json", "--api-url", "http://127.0.0.1:0"}, &output)
	if err != nil && !strings.Contains(err.Error(), "required checks failed") {
		t.Fatalf("doctor error = %v", err)
	}
	var report doctorReport
	if decodeErr := json.Unmarshal(output.Bytes(), &report); decodeErr != nil {
		t.Fatalf("decode doctor JSON: %v\n%s", decodeErr, output.String())
	}
	if len(report.Checks) < 5 {
		t.Fatalf("checks = %d, want prerequisite and resource checks", len(report.Checks))
	}
	for _, check := range report.Checks {
		if check.Name == "" || check.Status == "" {
			t.Fatalf("incomplete check: %#v", check)
		}
	}
}

func TestDoctorRejectsUnknownOutputFormat(t *testing.T) {
	var output bytes.Buffer
	if err := doctorWithIO([]string{"--output", "yaml"}, &output); err == nil {
		t.Fatal("doctor accepted unsupported output format")
	}
}

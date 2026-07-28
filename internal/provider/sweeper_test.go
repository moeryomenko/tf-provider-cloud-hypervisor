// Package provider contains the Terraform provider implementation and its
// acceptance test suite. This file implements sweepers for cleaning up
// cloud-hypervisor processes leaked by interrupted acceptance tests.
package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMain wraps resource.TestMain to add -sweep flag support. Running
// `go test -sweep= ./internal/provider/` triggers all registered sweepers,
// which clean up leaked cloud-hypervisor processes from failed or
// interrupted acceptance tests.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// init registers the cloudhypervisor_vm sweeper. The sweeper is triggered
// by the -sweep flag passed to go test (via TestMain/resource.TestMain).
func init() {
	resource.AddTestSweepers("cloudhypervisor_vm", &resource.Sweeper{
		Name: "cloudhypervisor_vm",
		F:    sweepCloudHypervisorVMs,
	})
}

// sweepCloudHypervisorVMs finds and kills cloud-hypervisor processes that
// were left running by interrupted acceptance tests.
//
// Safety: the function only targets processes whose --api-socket-path
// contains a test-identifying directory prefix ("ch-tf-"), preventing
// accidental termination of production cloud-hypervisor instances.
//
// For each matching process:
//  1. Send SIGTERM for graceful shutdown
//  2. Wait up to 5 seconds for the process to exit
//  3. If still running, send SIGKILL
//  4. Reap the process via syscall.Wait4
//  5. Remove the socket directory
//
// The region parameter is unused because cloud-hypervisor is a local-only
// VMM with no multi-region concept.
func sweepCloudHypervisorVMs(region string) error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("read /proc: %w", err)
	}

	// Collect all matching PIDs first so we don't miss any if earlier
	// cleanup operations interfere with the directory listing.
	type match struct {
		pid       int
		socketDir string
	}
	var matches []match

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pidStr := entry.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue // not a process directory
		}

		cmdline, err := os.ReadFile(filepath.Join("/proc", pidStr, "cmdline"))
		if err != nil {
			continue // process may have exited between readdir and readfile
		}

		// cmdline is NUL-separated; replace with spaces for string matching.
		cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
		if !strings.Contains(cmdStr, "cloud-hypervisor") {
			continue
		}

		// Safety: only match processes whose --api-socket-path contains
		// the test directory prefix. This prevents killing production
		// cloud-hypervisor instances.
		socketDir := extractSocketDir(cmdStr)
		if socketDir == "" || !isTestDir(socketDir) {
			continue
		}

		matches = append(matches, match{pid: pid, socketDir: socketDir})
	}

	if len(matches) == 0 {
		return nil
	}

	var errs []string
	for _, m := range matches {
		if err := killProcess(m.pid, m.socketDir); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("sweep errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// killProcess sends SIGTERM to the given PID, waits up to 5 seconds for
// graceful shutdown, then sends SIGKILL if still running. It reaps the
// process via syscall.Wait4 and removes the socket directory.
func killProcess(pid int, socketDir string) error {
	// Send SIGTERM for graceful shutdown.
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal SIGTERM to PID %d: %w", pid, err)
	}

	// Wait up to 5 seconds for the process to exit.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err != nil {
			// Process has exited — reap it and clean up.
			var wstatus syscall.WaitStatus
			if _, err := syscall.Wait4(pid, &wstatus, 0, nil); err != nil {
				return fmt.Errorf("wait4 PID %d: %w", pid, err)
			}
			_ = os.RemoveAll(socketDir)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill.
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("signal SIGKILL to PID %d: %w", pid, err)
	}

	// Reap the killed process.
	var wstatus syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &wstatus, 0, nil); err != nil {
		return fmt.Errorf("wait4 PID %d (after SIGKILL): %w", pid, err)
	}

	// Clean up the socket directory.
	_ = os.RemoveAll(socketDir)

	return nil
}

// extractSocketDir parses the process cmdline (with NUL bytes replaced by
// spaces) for the --api-socket-path argument and returns the parent
// directory of the socket file.
//
// Example cmdline input (after NUL replacement):
//
//	"cloud-hypervisor --api-socket-path /tmp/ch-tf-abc123/api.sock"
//
// Returns: "/tmp/ch-tf-abc123"
func extractSocketDir(cmdline string) string {
	// Split on the flag name. The value appears in the next token.
	parts := strings.Split(cmdline, "--api-socket-path")
	if len(parts) < 2 {
		return ""
	}

	// The value follows the flag, possibly after whitespace.
	rest := strings.TrimSpace(parts[1])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}

	return filepath.Dir(fields[0])
}

// isTestDir returns true if the directory path contains the test prefix
// used by acceptance test helpers to identify temp directories.
//
// The chproc.Manager creates temp dirs with prefix "ch-tf-*", and
// testAccExternalCHHelper creates dirs with prefix "ch-tf-ext-*".
// Both contain "ch-tf-" in the path, which serves as the common
// identifying prefix.
func isTestDir(dir string) bool {
	return strings.Contains(dir, "ch-tf-")
}

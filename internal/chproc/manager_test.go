// Package chproc_test contains tests for the cloud-hypervisor process manager.
//
// These tests use a fake-cloud-hypervisor binary fixture (in testdata/) and
// never require a real Cloud-Hypervisor VMM, KVM, or root permissions.
//
// RED PHASE: These tests reference types from package chproc that do not exist
// yet. They will fail to compile. This is the expected TDD red state.
package chproc_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/chproc"
)

// findFakeBinary resolves the absolute path to the fake-cloud-hypervisor
// fixture and verifies it is executable.
func findFakeBinary(t *testing.T) string {
	t.Helper()

	rel := filepath.Join("testdata", "fake-cloud-hypervisor")
	abs, err := filepath.Abs(rel)
	if err != nil {
		t.Fatalf("failed to resolve fake binary path %s: %v", rel, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		t.Fatalf("fake binary not found at %s: %v", abs, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("fake binary %s is not executable", abs)
	}
	return abs
}

// writeWrapper creates an executable shell script that invokes the real fake
// binary with all Manager-supplied args plus extra trailing flags.
func writeWrapper(t *testing.T, dir, name, fakeBin string, extraFlags ...string) string {
	t.Helper()

	var flagPart string
	if len(extraFlags) > 0 {
		flagPart = " " + strings.Join(extraFlags, " ")
	}

	script := fmt.Sprintf(`#!/bin/bash
exec "%s" "$@"%s
`, fakeBin, flagPart)

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write wrapper %s: %v", path, err)
	}
	return path
}

// TestStartStops starts the fake binary, verifies the socket path is returned
// and exists on disk, then stops the process and confirms cleanup.
func TestStartStops(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(fakeBin),
	)
	ctx := t.Context()

	socketPath, err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start() returned unexpected error: %v", err)
	}
	if socketPath == "" {
		t.Fatal("Start() returned empty socket path")
	}

	pid := mgr.PID()
	if pid == 0 {
		t.Error("PID() = 0 after Start, want non-zero")
	}

	if _, err := os.Stat(socketPath); err != nil {
		t.Errorf("socket file does not exist at %s: %v", socketPath, err)
	}

	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop() returned unexpected error: %v", err)
	}

	// Signal 0 tests whether the process exists without sending a signal.
	if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
		t.Error("process is still alive after Stop")
	}
}

// TestWaitReady starts the fake binary, then verifies that WaitReady succeeds
// once the socket appears.
func TestWaitReady(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(fakeBin),
	)
	ctx := t.Context()

	socketPath, err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// The fake binary creates the socket immediately (no --no-socket).
	// WaitReady with a generous timeout must succeed.
	if err := mgr.WaitReady(ctx, 5*time.Second); err != nil {
		t.Fatalf("WaitReady() returned error: %v", err)
	}

	if _, err := os.Stat(socketPath); err != nil {
		t.Errorf("socket not found after WaitReady: %v", err)
	}

	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}

// TestWaitReadyTimeout verifies that WaitReady returns a timeout error when
// the binary never creates the API socket.
func TestWaitReadyTimeout(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)

	// Create a wrapper that adds --no-socket so the fixture never creates
	// a listening socket. Manager passes "--api-socket-path <path>"; our
	// wrapper appends "--no-socket" at the end.
	noSocketBin := writeWrapper(t, t.TempDir(), "fake-no-socket", fakeBin,
		"--no-socket")

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(noSocketBin),
	)
	ctx := t.Context()

	if _, err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// WaitReady with a very short timeout must time out.
	if err := mgr.WaitReady(ctx, 500*time.Millisecond); err == nil {
		t.Error("WaitReady() succeeded but expected timeout error")
	} else {
		t.Logf("WaitReady expectedly failed with: %v", err)
	}

	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}

// TestStopDeadProcess verifies that Stop returns nil when called on a process
// that has already exited (idempotent cleanup).
func TestStopDeadProcess(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(fakeBin),
	)
	ctx := t.Context()

	_, err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Kill the process externally before calling Stop.
	pid := mgr.PID()
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatalf("failed to kill process %d: %v", pid, err)
	}

	// Wait for the process to fully die.
	for {
		var ws syscall.WaitStatus
		_, err := syscall.Wait4(pid, &ws, syscall.WNOHANG, nil)
		if errors.Is(err, syscall.ECHILD) {
			break // Child already reaped.
		}
		if err != nil {
			t.Fatalf("Wait4 failed: %v", err)
		}
		if ws.Exited() || ws.Signaled() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Stop should succeed even though the process is already dead.
	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop() on dead process returned error: %v", err)
	}
}

// TestNoZombie verifies that after Stop, the child process is reaped (not a
// zombie). Uses syscall.Wait4 with WNOHANG to confirm no child remains.
func TestNoZombie(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(fakeBin),
	)
	ctx := t.Context()

	_, err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	pid := mgr.PID()

	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// After Stop, the process must be reaped. Wait4 with WNOHANG should
	// return ECHILD (no such child).
	_, err = syscall.Wait4(pid, nil, syscall.WNOHANG, nil)
	if !errors.Is(err, syscall.ECHILD) {
		t.Errorf("process %d not reaped: Wait4=%v (want ECHILD)", pid, err)
	}
}

// TestTempDirCleanup verifies that the temp directory created by Start is
// removed by Stop.
func TestTempDirCleanup(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)
	socketDir := t.TempDir()

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(fakeBin),
		chproc.WithSocketDir(socketDir),
	)
	ctx := t.Context()

	_, err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if err := mgr.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if _, err := os.Stat(socketDir); err == nil {
		t.Error("socket directory still exists after Stop")
	} else if !os.IsNotExist(err) {
		t.Errorf("unexpected stat error for socket dir: %v", err)
	}
}

// TestStderrCapture verifies that when the binary exits early with a non-zero
// exit code and an error message on stderr, Start returns an error containing
// the stderr output.
func TestStderrCapture(t *testing.T) {
	t.Parallel()

	fakeBin := findFakeBinary(t)
	expectedMsg := "FAIL: invalid configuration"
	earlyExitBin := writeWrapper(t, t.TempDir(), "fake-early-exit", fakeBin,
		"--exit-on-start", "--exit-code", "1", "--stderr-msg", expectedMsg)

	mgr := chproc.NewManager(
		chproc.WithBinaryPath(earlyExitBin),
	)
	ctx := t.Context()

	_, err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start() succeeded but expected error from early-exiting binary")
	}

	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Start() error %q does not contain stderr message %q",
			err.Error(), expectedMsg)
	}
}

// TestMissingBinary verifies that Start returns an error when the configured
// binary path does not exist.
func TestMissingBinary(t *testing.T) {
	t.Parallel()

	mgr := chproc.NewManager(
		chproc.WithBinaryPath("/nonexistent/cloud-hypervisor"),
	)
	ctx := t.Context()

	_, err := mgr.Start(ctx)
	if err == nil {
		t.Fatal("Start() succeeded but expected error for missing binary")
	}

	t.Logf("Start() correctly returned error: %v", err)
}

// Package chproc manages the cloud-hypervisor binary process lifecycle.
//
// It handles starting the binary, waiting for its API socket to become
// reachable, and graceful/forced shutdown with cleanup.
package chproc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// Manager manages a cloud-hypervisor process lifecycle.
type Manager struct {
	binaryPath string
	socketDir  string
	socketPath string

	cmd    *exec.Cmd
	cancel context.CancelFunc

	stderrBuf  bytes.Buffer
	stderrDone chan struct{}
	waitDone   chan struct{}
	waitErr    error
}

// NewManager creates a new process manager with default settings.
// Defaults:
//   - binaryPath: "cloud-hypervisor" (resolved via exec.LookPath)
//   - socketDir: auto-created via os.MkdirTemp (prefix "ch-tf-*")
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		binaryPath: "cloud-hypervisor",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithBinaryPath sets the path or name of the cloud-hypervisor binary.
func WithBinaryPath(path string) ManagerOption {
	return func(m *Manager) {
		m.binaryPath = path
	}
}

// WithSocketDir sets an explicit socket directory instead of auto-creating one.
// The directory must already exist when Start is called.
func WithSocketDir(path string) ManagerOption {
	return func(m *Manager) {
		m.socketDir = path
	}
}

// Start starts the cloud-hypervisor binary with --api-socket-path pointing to
// a file inside the socket directory. It waits for the process to begin
// executing but does NOT wait for the socket to become reachable (use
// WaitReady for that).
//
// The context controls startup only: cancelling the context will kill the
// process and clean up. The returned process outlives the context otherwise.
//
// If the binary exits prematurely (within ~500ms of starting), Start returns
// an error containing any captured stderr output.
//
// Start is idempotent when called multiple times: subsequent calls return the
// existing socket path.
func (m *Manager) Start(ctx context.Context) (socketPath string, err error) {
	if m.cmd != nil && m.cmd.Process != nil {
		return m.socketPath, nil
	}

	binaryPath, err := exec.LookPath(m.binaryPath)
	if err != nil {
		return "", fmt.Errorf("cloud-hypervisor binary not found: %w", err)
	}

	dir := m.socketDir
	if dir == "" {
		dir, err = os.MkdirTemp("", "ch-tf-*")
		if err != nil {
			return "", fmt.Errorf("create temp socket dir: %w", err)
		}
		m.socketDir = dir
	} else {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create socket dir: %w", err)
		}
	}

	socketPath = filepath.Join(dir, "api.sock")
	m.socketPath = socketPath

	cmd := exec.Command(binaryPath, "--api-socket", "path="+socketPath)
	m.cmd = cmd

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("create stderr pipe: %w", err)
	}

	m.stderrBuf.Reset()
	m.stderrDone = make(chan struct{})
	go func() {
		_, _ = io.Copy(&m.stderrBuf, stderrPipe)
		close(m.stderrDone)
	}()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start cloud-hypervisor: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	m.waitDone = make(chan struct{})
	go func() {
		m.waitErr = cmd.Wait()
		close(m.waitDone)
	}()

	// Wait briefly to detect immediate crash. If the process exits within
	// 500ms, treat it as a startup failure and return the error with any
	// captured stderr.
	select {
	case <-m.waitDone:
		<-m.stderrDone
		stderrOutput := m.stderrBuf.String()
		if stderrOutput != "" {
			return "", fmt.Errorf("cloud-hypervisor exited prematurely: %s: %w",
				stderrOutput, m.waitErr)
		}
		return "", fmt.Errorf("cloud-hypervisor exited prematurely: %w", m.waitErr)

	case <-time.After(500 * time.Millisecond):
		return socketPath, nil

	case <-ctx.Done():
		_ = m.cmd.Process.Signal(syscall.SIGKILL)
		<-m.waitDone
		cancel()
		return "", ctx.Err()
	}
}

// WaitReady polls the API socket until it is reachable or the timeout expires.
// Uses exponential backoff (50ms, 100ms, 200ms, capped at 500ms) between
// polls. Returns nil when the socket accepts a connection.
//
// If the process has exited before the socket becomes reachable, WaitReady
// returns a process-exited error immediately.
//
// If the context is cancelled, returns the context error.
func (m *Manager) WaitReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	backoff := 50 * time.Millisecond
	const maxBackoff = 500 * time.Millisecond

	for {
		// Check if the process has exited.
		select {
		case <-m.waitDone:
			<-m.stderrDone
			stderrOutput := m.stderrBuf.String()
			if stderrOutput != "" {
				return fmt.Errorf("cloud-hypervisor exited before socket ready: %s: %w",
					stderrOutput, m.waitErr)
			}
			return fmt.Errorf("cloud-hypervisor exited before socket ready: %w", m.waitErr)
		default:
		}

		conn, err := net.DialTimeout("unix", m.socketPath, backoff)
		if err == nil {
			conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}

// Stop sends SIGTERM to the child process, waits up to 5 seconds for it to
// exit, then sends SIGKILL if still running. It waits for the process to be
// reaped and cleans up the temp socket directory.
//
// Idempotent: calling Stop on an already-stopped manager returns nil.
// Safe to call if Start was never called (returns nil).
func (m *Manager) Stop(_ context.Context) error {
	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}

	// Check if the process has already exited.
	select {
	case <-m.waitDone:
		m.cleanup()
		return nil
	default:
	}

	// Try graceful shutdown.
	_ = m.cmd.Process.Signal(syscall.SIGTERM)

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-m.waitDone:
		m.cleanup()
		return nil
	case <-timer.C:
		_ = m.cmd.Process.Signal(syscall.SIGKILL)
		<-m.waitDone
		m.cleanup()
		return nil
	}
}

// PID returns the process ID of the child. Returns 0 if Start has not been
// called or the process has exited.
func (m *Manager) PID() int {
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Pid
	}
	return 0
}

func (m *Manager) cleanup() {
	if m.socketDir != "" {
		os.RemoveAll(m.socketDir)
	}
	if m.cancel != nil {
		m.cancel()
	}
}

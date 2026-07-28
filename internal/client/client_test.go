package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/client"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// requestRecorder is a test HTTP handler that records the last request and
// responds with a configurable status code and body. The handler must be
// wrapped with StripPrefix("/api/v1", ...) so that req.URL.Path reflects the
// endpoint path (e.g. /vmm.ping) without the /api/v1 prefix.
type requestRecorder struct {
	mu      sync.Mutex
	method  string
	path    string
	body    []byte
	headers http.Header

	status int
	resp   string
}

func newRecorder(status int, resp string) *requestRecorder {
	return &requestRecorder{status: status, resp: resp}
}

func (r *requestRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.method = req.Method
	r.path = req.URL.Path
	r.headers = req.Header.Clone()

	data, _ := io.ReadAll(req.Body)
	r.body = data
	req.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	if r.resp != "" {
		fmt.Fprint(w, r.resp)
	}
}

// newUDSServer starts an HTTP server on a Unix domain socket and returns the
// socket path and a cleanup function. The handler is wrapped with
// StripPrefix("/api/v1", ...) so that recorded paths match endpoint names.
func newUDSServer(t *testing.T, handler http.Handler) (socketPath string, cleanup func()) {
	t.Helper()

	dir := t.TempDir()
	socketPath = filepath.Join(dir, "api.sock")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on unix socket %s: %v", socketPath, err)
	}

	srv := &http.Server{
		Handler: http.StripPrefix("/api/v1", handler),
	}
	go func() {
		srv.Serve(listener) //nolint:errcheck
	}()

	cleanup = func() { srv.Close() }
	return socketPath, cleanup
}

// newHTTPServer starts an httptest.Server and returns a base URL
// (http://host:port/api/v1) plus a cleanup wrapper. The handler is wrapped
// with StripPrefix so recorded paths match endpoint names (no /api/v1 prefix).
func newHTTPServer(t *testing.T, handler http.Handler) (baseURL string, cleanup func()) {
	t.Helper()
	srv := httptest.NewServer(http.StripPrefix("/api/v1", handler))
	return srv.URL + "/api/v1", srv.Close
}

// ---------------------------------------------------------------------------
// Constructor & Option Tests
// ---------------------------------------------------------------------------

func TestNew_UDS(t *testing.T) {
	t.Parallel()

	rec := newRecorder(http.StatusOK, `{"version":"1.0","pid":42}`)
	socketPath, cleanup := newUDSServer(t, rec)
	defer cleanup()

	c, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("New(%q) unexpected error: %v", socketPath, err)
	}
	if c == nil {
		t.Fatal("New returned nil client")
	}

	// Verify the client can reach the server by making a Ping call.
	resp, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping after New failed: %v", err)
	}
	if resp.Version != "1.0" {
		t.Errorf("Ping.Version = %q, want %q", resp.Version, "1.0")
	}
}

func TestNew_UDS_dialError(t *testing.T) {
	t.Parallel()

	// Non‑existent socket — New may eagerly validate, or Ping may fail.
	// Accept either so the implementation can choose eager vs lazy dial.
	socketPath := "/tmp/nonexistent-test-socket-XXXXXXXX.sock"
	c, err := client.New(socketPath)

	if err == nil && c != nil {
		// Lazy dial — first API call must fail.
		_, pingErr := c.Ping(context.Background())
		if pingErr == nil {
			t.Fatal("expected error from Ping on non‑existent socket, got nil")
		}
	} else if err == nil {
		t.Fatal("New returned nil client without error on non‑existent socket")
	}
	// err != nil is fine (eager validation).
}

func TestNewHTTP_TCP(t *testing.T) {
	t.Parallel()

	rec := newRecorder(http.StatusOK, `{"version":"1.0","pid":42}`)
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatalf("NewHTTP(%q) unexpected error: %v", baseURL, err)
	}
	if c == nil {
		t.Fatal("NewHTTP returned nil client")
	}

	resp, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping after NewHTTP failed: %v", err)
	}
	if resp.Version != "1.0" {
		t.Errorf("Ping.Version = %q, want %q", resp.Version, "1.0")
	}
}

func TestNewHTTP_emptyBaseURL(t *testing.T) {
	t.Parallel()

	c, err := client.NewHTTP("")
	if err == nil {
		t.Fatal("expected error for empty base URL, got nil")
	}
	if c != nil {
		t.Fatal("expected nil client on empty base URL")
	}
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	rec := newRecorder(http.StatusOK, `{"version":"1.0"}`)
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL, client.WithTimeout(100*time.Millisecond))
	if err != nil {
		t.Fatalf("NewHTTP with WithTimeout: %v", err)
	}
	if c == nil {
		t.Fatal("NewHTTP returned nil client")
	}

	resp, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if resp.Version != "1.0" {
		t.Errorf("Ping.Version = %q, want %q", resp.Version, "1.0")
	}
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

func TestClient_Ping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		response  string
		wantErr   bool
		wantErrIs error
		wantVer   string
		wantPID   int64
		wantFeat  int
	}{
		{
			name:     "success",
			status:   http.StatusOK,
			response: `{"version":"0.3.0","build_version":"v0.3.0","pid":12345,"features":["acpi","tdx"]}`,
			wantVer:  "0.3.0",
			wantPID:  12345,
			wantFeat: 2,
		},
		{
			name:     "minimal_response",
			status:   http.StatusOK,
			response: `{"version":"1.0"}`,
			wantVer:  "1.0",
		},
		{
			name:     "malformed_json",
			status:   http.StatusOK,
			response: `{"version":broken}`,
			wantErr:  true,
		},
		{
			name:     "server_error_500",
			status:   http.StatusInternalServerError,
			response: `{"error":"internal fault"}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := c.Ping(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Version != tt.wantVer {
				t.Errorf("Version = %q, want %q", resp.Version, tt.wantVer)
			}
			if resp.PID != tt.wantPID {
				t.Errorf("PID = %d, want %d", resp.PID, tt.wantPID)
			}
			if len(resp.Features) != tt.wantFeat {
				t.Errorf("len(Features) = %d, want %d", len(resp.Features), tt.wantFeat)
			}

			// Verify request method and path.
			if rec.method != http.MethodGet {
				t.Errorf("method = %q, want %q", rec.method, http.MethodGet)
			}
			if rec.path != "/vmm.ping" {
				t.Errorf("path = %q, want %q", rec.path, "/vmm.ping")
			}
		})
	}
}

func TestClient_Ping_UDS(t *testing.T) {
	t.Parallel()

	rec := newRecorder(http.StatusOK, `{"version":"0.3.0","pid":99}`)
	socketPath, cleanup := newUDSServer(t, rec)
	defer cleanup()

	c, err := client.New(socketPath)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Version != "0.3.0" {
		t.Errorf("Version = %q, want %q", resp.Version, "0.3.0")
	}

	if rec.method != http.MethodGet {
		t.Errorf("method = %q, want %q", rec.method, http.MethodGet)
	}
	if rec.path != "/vmm.ping" {
		t.Errorf("path = %q, want %q", rec.path, "/vmm.ping")
	}
}

// ---------------------------------------------------------------------------
// VMInfo
// ---------------------------------------------------------------------------

func TestClient_VMInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		response  string
		wantErr   bool
		wantErrIs error
		wantState string
	}{
		{
			name:   "success",
			status: http.StatusOK,
			response: `{
				"config": {"payload":{"kernel":"/vmlinuz","cmdline":"console=hvc0"}},
				"state": "Created"
			}`,
			wantState: "Created",
		},
		{
			name:   "running",
			status: http.StatusOK,
			response: `{
				"config": {"payload":{"kernel":"/vmlinuz"}},
				"state": "Running"
			}`,
			wantState: "Running",
		},
		{
			name:      "shutdown",
			status:    http.StatusOK,
			response:  `{"config":{"payload":{"kernel":"/vmlinuz"}},"state":"Shutdown"}`,
			wantState: "Shutdown",
		},
		{
			name:      "not_found",
			status:    http.StatusNotFound,
			response:  `{"error":"VM not created"}`,
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			info, err := c.VMInfo(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info == nil {
				t.Fatal("VMInfo returned nil")
			}
			if string(info.State) != tt.wantState {
				t.Errorf("State = %q, want %q", info.State, tt.wantState)
			}
			if info.Config == nil {
				t.Error("Config is nil")
			}

			// Verify request.
			if rec.method != http.MethodGet {
				t.Errorf("method = %q, want %q", rec.method, http.MethodGet)
			}
			if rec.path != "/vm.info" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.info")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CreateVM
// ---------------------------------------------------------------------------

func TestClient_CreateVM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		response  string
		cfg       *client.VmConfig
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "success",
			status: http.StatusNoContent,
			cfg: &client.VmConfig{
				Payload: &client.PayloadConfig{Kernel: "/vmlinuz"},
			},
		},
		{
			name:   "minimal_config",
			status: http.StatusNoContent,
			cfg: &client.VmConfig{
				Payload: &client.PayloadConfig{
					Kernel:  "/vmlinuz",
					Cmdline: "console=hvc0",
				},
			},
		},
		{
			name:   "with_cpus_and_memory",
			status: http.StatusNoContent,
			cfg: &client.VmConfig{
				Payload: &client.PayloadConfig{Kernel: "/vmlinuz"},
				Cpus: &client.CpusConfig{
					BootVcpus: 2,
					MaxVcpus:  4,
				},
				Memory: &client.MemoryConfig{
					Size: 512_000_000,
				},
			},
		},
		{
			name:     "server_rejected",
			status:   http.StatusBadRequest,
			response: `{"error":"invalid config"}`,
			cfg:      &client.VmConfig{Payload: &client.PayloadConfig{Kernel: "/vmlinuz"}},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = c.CreateVM(context.Background(), tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify request details.
			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.create" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.create")
			}
			if ct := rec.headers.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ct, "application/json")
			}

			// Verify body is valid JSON that decodes to a VmConfig with payload.
			if len(rec.body) == 0 {
				t.Fatal("request body is empty")
			}
			var got client.VmConfig
			if err := json.Unmarshal(rec.body, &got); err != nil {
				t.Fatalf("request body is not valid JSON: %v", err)
			}
			if got.Payload == nil {
				t.Error("request body missing payload field")
			}
		})
	}
}

func TestClient_CreateVM_AlreadyExists(t *testing.T) {
	t.Parallel()

	// 405 returned when VM already exists.
	rec := newRecorder(http.StatusMethodNotAllowed, `{"error":"VM already created"}`)
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &client.VmConfig{Payload: &client.PayloadConfig{Kernel: "/vmlinuz"}}
	err = c.CreateVM(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, client.ErrInvalidState) {
		t.Errorf("error does not wrap ErrInvalidState: %v", err)
	}
}

// ---------------------------------------------------------------------------
// BootVM
// ---------------------------------------------------------------------------

func TestClient_BootVM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		response  string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "success",
			status: http.StatusNoContent,
		},
		{
			name:      "not_created",
			status:    http.StatusNotFound,
			response:  `{"error":"VM not created"}`,
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
		{
			name:      "already_booted",
			status:    http.StatusMethodNotAllowed,
			response:  `{"error":"not in Created state"}`,
			wantErr:   true,
			wantErrIs: client.ErrInvalidState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = c.BootVM(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.boot" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.boot")
			}
			// BootVM has no request body.
			if len(rec.body) != 0 {
				t.Errorf("expected empty body, got %d bytes", len(rec.body))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ShutdownVM
// ---------------------------------------------------------------------------

func TestClient_ShutdownVM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		response  string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "success",
			status: http.StatusNoContent,
		},
		{
			name:      "not_created",
			status:    http.StatusNotFound,
			response:  `{"error":"VM not created"}`,
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
		{
			name:      "not_started",
			status:    http.StatusMethodNotAllowed,
			response:  `{"error":"VM not started"}`,
			wantErr:   true,
			wantErrIs: client.ErrInvalidState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = c.ShutdownVM(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.shutdown" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.shutdown")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteVM
// ---------------------------------------------------------------------------

func TestClient_DeleteVM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		response  string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "success",
			status: http.StatusNoContent,
		},
		{
			name:      "not_found",
			status:    http.StatusNotFound,
			response:  `{"error":"not found"}`,
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = c.DeleteVM(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.delete" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.delete")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AddDisk
// ---------------------------------------------------------------------------

func TestClient_AddDisk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		response   string
		cfg        *client.DiskConfig
		wantErr    bool
		wantErrIs  error
		wantID     string
		wantBDF    string
		wantNilPCI bool
	}{
		{
			name:     "hotplug",
			status:   http.StatusOK,
			response: `{"id":"disk-1","bdf":"0000:00:03.0"}`,
			cfg:      &client.DiskConfig{Path: "/var/lib/disks/test.raw"},
			wantID:   "disk-1", wantBDF: "0000:00:03.0",
		},
		{
			name:       "cold_plug",
			status:     http.StatusNoContent,
			cfg:        &client.DiskConfig{Path: "/var/lib/disks/test.raw"},
			wantNilPCI: true,
		},
		{
			name:     "full_config",
			status:   http.StatusOK,
			response: `{"id":"disk-2","bdf":"0000:00:04.0"}`,
			cfg:      &client.DiskConfig{Path: "/disk.qcow2", ID: "disk-2"},
			wantID:   "disk-2", wantBDF: "0000:00:04.0",
		},
		{
			name:      "not_created",
			status:    http.StatusNotFound,
			response:  `{"error":"VM not created"}`,
			cfg:       &client.DiskConfig{Path: "/disk.raw"},
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
		{
			name:     "server_error",
			status:   http.StatusInternalServerError,
			response: `{"error":"disk error"}`,
			cfg:      &client.DiskConfig{Path: "/disk.raw"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			pci, err := c.AddDisk(context.Background(), tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNilPCI {
				if pci != nil {
					t.Errorf("expected nil PciDeviceInfo, got %+v", pci)
				}
			} else {
				if pci == nil {
					t.Fatal("expected non-nil PciDeviceInfo, got nil")
				}
				if pci.ID != tt.wantID {
					t.Errorf("PciDeviceInfo.ID = %q, want %q", pci.ID, tt.wantID)
				}
				if pci.BDF != tt.wantBDF {
					t.Errorf("PciDeviceInfo.BDF = %q, want %q", pci.BDF, tt.wantBDF)
				}
			}

			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.add-disk" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.add-disk")
			}
			if len(rec.body) == 0 {
				t.Fatal("request body is empty")
			}

			var sent client.DiskConfig
			if err := json.Unmarshal(rec.body, &sent); err != nil {
				t.Fatalf("request body not valid JSON: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AddNet
// ---------------------------------------------------------------------------

func TestClient_AddNet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		response   string
		cfg        *client.NetConfig
		wantErr    bool
		wantErrIs  error
		wantID     string
		wantBDF    string
		wantNilPCI bool
	}{
		{
			name:     "hotplug",
			status:   http.StatusOK,
			response: `{"id":"net-1","bdf":"0000:00:05.0"}`,
			cfg:      &client.NetConfig{Tap: "tap0", MAC: "de:ad:be:ef:00:01"},
			wantID:   "net-1", wantBDF: "0000:00:05.0",
		},
		{
			name:       "cold_plug",
			status:     http.StatusNoContent,
			cfg:        &client.NetConfig{Tap: "tap0"},
			wantNilPCI: true,
		},
		{
			name:      "not_created",
			status:    http.StatusNotFound,
			response:  `{"error":"not created"}`,
			cfg:       &client.NetConfig{Tap: "tap0"},
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			pci, err := c.AddNet(context.Background(), tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNilPCI {
				if pci != nil {
					t.Errorf("expected nil PciDeviceInfo, got %+v", pci)
				}
			} else {
				if pci == nil {
					t.Fatal("expected non-nil PciDeviceInfo, got nil")
				}
				if pci.ID != tt.wantID {
					t.Errorf("PciDeviceInfo.ID = %q, want %q", pci.ID, tt.wantID)
				}
				if pci.BDF != tt.wantBDF {
					t.Errorf("PciDeviceInfo.BDF = %q, want %q", pci.BDF, tt.wantBDF)
				}
			}

			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.add-net" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.add-net")
			}
			if len(rec.body) == 0 {
				t.Fatal("request body is empty")
			}

			var sent client.NetConfig
			if err := json.Unmarshal(rec.body, &sent); err != nil {
				t.Fatalf("request body not valid JSON: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RemoveDevice
// ---------------------------------------------------------------------------

func TestClient_RemoveDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		deviceID  string
		status    int
		response  string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "success",
			deviceID: "disk-1",
			status:   http.StatusNoContent,
		},
		{
			name:      "not_found",
			deviceID:  "nonexistent",
			status:    http.StatusNotFound,
			response:  `{"error":"no such device"}`,
			wantErr:   true,
			wantErrIs: client.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder(tt.status, tt.response)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = c.RemoveDevice(context.Background(), tt.deviceID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error does not wrap %v: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.method != http.MethodPut {
				t.Errorf("method = %q, want %q", rec.method, http.MethodPut)
			}
			if rec.path != "/vm.remove-device" {
				t.Errorf("path = %q, want %q", rec.path, "/vm.remove-device")
			}

			// Verify the body contains {"id":"<deviceID>"}.
			var body map[string]string
			if err := json.Unmarshal(rec.body, &body); err != nil {
				t.Fatalf("request body is not valid JSON: %v", err)
			}
			if body["id"] != tt.deviceID {
				t.Errorf("body.id = %q, want %q", body["id"], tt.deviceID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error Mapping
// ---------------------------------------------------------------------------

func TestClient_ErrorMapping_404(t *testing.T) {
	t.Parallel()

	endpoints := []struct {
		name string
		fn   func(ctx context.Context, c *client.Client) error
	}{
		{"VMInfo", func(ctx context.Context, c *client.Client) error {
			_, err := c.VMInfo(ctx)
			return err
		}},
		{"BootVM", func(ctx context.Context, c *client.Client) error {
			return c.BootVM(ctx)
		}},
		{"ShutdownVM", func(ctx context.Context, c *client.Client) error {
			return c.ShutdownVM(ctx)
		}},
		{"DeleteVM", func(ctx context.Context, c *client.Client) error {
			return c.DeleteVM(ctx)
		}},
		{"RemoveDevice", func(ctx context.Context, c *client.Client) error {
			return c.RemoveDevice(ctx, "test-id")
		}},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			rec := newRecorder(http.StatusNotFound, `{"error":"not found"}`)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = ep.fn(context.Background(), c)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, client.ErrNotFound) {
				t.Errorf("error does not wrap ErrNotFound: %v", err)
			}
		})
	}
}

func TestClient_ErrorMapping_405(t *testing.T) {
	t.Parallel()

	endpoints := []struct {
		name string
		fn   func(ctx context.Context, c *client.Client) error
	}{
		{"ShutdownVM", func(ctx context.Context, c *client.Client) error {
			return c.ShutdownVM(ctx)
		}},
		{"BootVM", func(ctx context.Context, c *client.Client) error {
			return c.BootVM(ctx)
		}},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			rec := newRecorder(http.StatusMethodNotAllowed, `{"error":"invalid state"}`)
			baseURL, cleanup := newHTTPServer(t, rec)
			defer cleanup()

			c, err := client.NewHTTP(baseURL)
			if err != nil {
				t.Fatal(err)
			}

			err = ep.fn(context.Background(), c)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, client.ErrInvalidState) {
				t.Errorf("error does not wrap ErrInvalidState: %v", err)
			}
		})
	}
}

func TestClient_ErrorMapping_500(t *testing.T) {
	t.Parallel()

	rec := newRecorder(http.StatusInternalServerError, `{"error":"internal server error"}`)
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// The error should contain the HTTP status code so callers can inspect it.
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code 500: %v", err)
	}
}

func TestClient_ErrorMapping_ConnectionRefused(t *testing.T) {
	t.Parallel()

	// Attempt to connect to a closed port — should produce a dial/refused error.
	c, err := client.NewHTTP("http://127.0.0.1:1/api/v1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}

	// The error should wrap a lower-level network error.
	var netErr net.Error
	if !errors.As(err, &netErr) {
		t.Logf("note: error does not wrap net.Error, but got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Context Cancellation & Timeout
// ---------------------------------------------------------------------------

func TestClient_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Server that blocks until the client cancels.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	baseURL, cleanup := newHTTPServer(t, handler)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = c.Ping(ctx)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error does not wrap context.Canceled: %v", err)
	}
}

func TestClient_ContextTimeout(t *testing.T) {
	t.Parallel()

	// Server that delays longer than the context deadline.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		fmt.Fprint(w, `{"version":"1.0"}`)
	})
	baseURL, cleanup := newHTTPServer(t, handler)
	defer cleanup()

	c, err := client.NewHTTP(baseURL, client.WithTimeout(10*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = c.Ping(ctx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error does not wrap context.DeadlineExceeded: %v", err)
	}
}

// ---------------------------------------------------------------------------
// JSON Serialization
// ---------------------------------------------------------------------------

func TestClient_JSONSerialization_SnakeCase(t *testing.T) {
	t.Parallel()

	cfg := &client.VmConfig{
		Payload: &client.PayloadConfig{
			Kernel:    "/vmlinuz",
			Initramfs: "/initrd.img",
		},
		Cpus: &client.CpusConfig{
			BootVcpus: 2,
			MaxVcpus:  4,
		},
		Memory: &client.MemoryConfig{
			Size: 512_000_000,
		},
	}

	rec := newRecorder(http.StatusNoContent, "")
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.CreateVM(context.Background(), cfg); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}

	// Parse request body and verify snake_case keys.
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.body, &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}

	// Top-level keys must be snake_case.
	for _, key := range []string{"payload", "cpus", "memory"} {
		if _, ok := body[key]; !ok {
			t.Errorf("missing snake_case key %q in request body", key)
		}
	}

	// CamelCase keys must NOT appear.
	for _, key := range []string{"Payload", "Cpus", "Memory"} {
		if _, ok := body[key]; ok {
			t.Errorf("unexpected camelCase key %q in request body (should be snake_case)", key)
		}
	}

	// Check nested snake_case in cpus.
	var cpusBody map[string]json.RawMessage
	if err := json.Unmarshal(body["cpus"], &cpusBody); err != nil {
		t.Fatalf("unmarshal cpus: %v", err)
	}
	for _, key := range []string{"boot_vcpus", "max_vcpus"} {
		if _, ok := cpusBody[key]; !ok {
			t.Errorf("cpus missing snake_case key %q", key)
		}
	}
}

func TestClient_JSONSerialization_OmitEmpty(t *testing.T) {
	t.Parallel()

	// Minimal VmConfig — only required fields should be present.
	cfg := &client.VmConfig{
		Payload: &client.PayloadConfig{Kernel: "/vmlinuz"},
	}

	rec := newRecorder(http.StatusNoContent, "")
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.CreateVM(context.Background(), cfg); err != nil {
		t.Fatalf("CreateVM: %v", err)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.body, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	// Optional top-level fields must be absent (omitempty).
	for _, key := range []string{"cpus", "memory", "disks", "net", "rng", "balloon",
		"fs", "pmem", "serial", "console", "devices", "vsock",
		"numa", "iommu", "watchdog", "rtc", "platform", "tpm",
		"vdpa", "user_devices", "pci_segments", "landlock_enable"} {
		if _, ok := body[key]; ok {
			t.Errorf("optional key %q should be omitted (omitempty)", key)
		}
	}

	// Required key must be present.
	if _, ok := body["payload"]; !ok {
		t.Error("required key 'payload' is missing from request body")
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestClient_EmptyBody_204(t *testing.T) {
	t.Parallel()

	// 204 No Content with empty body — client must not panic or block.
	rec := newRecorder(http.StatusNoContent, "")
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	// CreateVM returns 204 with no body.
	if err := c.CreateVM(context.Background(), &client.VmConfig{
		Payload: &client.PayloadConfig{Kernel: "/vmlinuz"},
	}); err != nil {
		t.Fatalf("CreateVM with empty 204 body: %v", err)
	}

	// Also verify BootVM, ShutdownVM, DeleteVM handle empty 204.
	if err := c.BootVM(context.Background()); err != nil {
		t.Fatalf("BootVM with empty 204 body: %v", err)
	}
	if err := c.ShutdownVM(context.Background()); err != nil {
		t.Fatalf("ShutdownVM with empty 204 body: %v", err)
	}
}

func TestClient_MalformedJSON(t *testing.T) {
	t.Parallel()

	rec := newRecorder(http.StatusOK, `this is not json at all`)
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// Must be a JSON syntax error.
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("error should be a json.SyntaxError, got: %T %v", err, err)
	}
}

func TestClient_VMInfo_ReadOnlyFields(t *testing.T) {
	t.Parallel()

	// VmInfo response with optional read-only fields populated by the server.
	rec := newRecorder(http.StatusOK, `{
		"config":{"payload":{"kernel":"/vmlinuz"}},
		"state":"Running",
		"memory_actual_size":1073741824
	}`)
	baseURL, cleanup := newHTTPServer(t, rec)
	defer cleanup()

	c, err := client.NewHTTP(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	info, err := c.VMInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Config == nil || info.Config.Payload == nil {
		t.Fatal("Config.Payload is nil")
	}
	if info.Config.Payload.Kernel != "/vmlinuz" {
		t.Errorf("Kernel = %q, want %q", info.Config.Payload.Kernel, "/vmlinuz")
	}
}

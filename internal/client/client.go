package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// DefaultTimeout is the default HTTP client timeout for CH API calls.
const DefaultTimeout = 30 * time.Second

// ---------------------------------------------------------------------------
// Option
// ---------------------------------------------------------------------------

// Option configures a Client.
type Option func(*options)

type options struct {
	timeout time.Duration
}

// WithTimeout sets the HTTP client timeout. If not set, DefaultTimeout is used.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client is an HTTP client for the Cloud-Hypervisor REST API.
//
// It supports both Unix domain socket (UDS) and TCP connections. Use New to
// connect via UDS and NewHTTP to connect via TCP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a Client connected to the CH API via a Unix domain socket.
//
// The socketPath is the path to the CH API Unix domain socket (e.g.,
// /tmp/ch-socket.sock). The client dials the socket directly and uses
// http://localhost as the HTTP host for request construction.
func New(socketPath string, opts ...Option) (*Client, error) {
	cfg := options{timeout: DefaultTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &Client{
		baseURL: "http://localhost/api/v1",
		httpClient: &http.Client{
			Timeout:   cfg.timeout,
			Transport: transport,
		},
	}, nil
}

// NewHTTP creates a Client connected to the CH API via TCP HTTP.
//
// The baseURL should be the full base URL including the /api/v1 path prefix
// (e.g., "http://localhost:8080/api/v1").
func NewHTTP(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL must not be empty")
	}

	cfg := options{timeout: DefaultTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: cfg.timeout,
		},
	}, nil
}

// ---------------------------------------------------------------------------
// doRequest — internal helper
// ---------------------------------------------------------------------------

// doRequest sends an HTTP request, checks the status code, and unmarshals the
// response body into responseTarget (if non-nil and status is 200).
//
// The boolean return value indicates whether the response body was decoded
// into responseTarget (true for 200, false for 204 No Content). Callers that
// need to distinguish hotplug (200 with body) from cold plug (204) should
// check this flag.
//
// Status code handling:
//   - 204 No Content — returns (false, nil) — no body expected
//   - 404 — returns (false, error wrapping ErrNotFound)
//   - 405 — returns (false, error wrapping ErrInvalidState)
//   - 4xx/5xx — returns (false, descriptive Error with status code and body)
//   - 200 — unmarshals JSON body into responseTarget, returns (true, nil)
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body, responseTarget any) (bool, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return false, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Network errors (dial refused, timeout, context cancelled) pass through.
		return false, err
	}
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode {
	case http.StatusNoContent:
		return false, nil

	case http.StatusNotFound:
		snippet := readBodySnippet(resp.Body)
		return false, fmt.Errorf("%w: %s", ErrNotFound, snippet)

	case http.StatusMethodNotAllowed:
		snippet := readBodySnippet(resp.Body)
		return false, fmt.Errorf("%w: %s", ErrInvalidState, snippet)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := strings.TrimSpace(string(bodyBytes))
		return false, &Error{
			StatusCode: resp.StatusCode,
			Body:       bodyStr,
		}
	}

	// 200 OK — decode response body if a target is provided.
	if responseTarget != nil {
		if err := json.NewDecoder(resp.Body).Decode(responseTarget); err != nil {
			return false, fmt.Errorf("decode response: %w", err)
		}
	}

	return true, nil
}

// readBodySnippet reads up to 512 bytes of the response body for error
// messages, then closes the body.
func readBodySnippet(r io.ReadCloser) string {
	if r == nil {
		return ""
	}
	defer r.Close() //nolint:errcheck
	data, _ := io.ReadAll(r)
	// Limit snippet length to avoid enormous error messages.
	const maxLen = 512
	if len(data) > maxLen {
		return strings.TrimSpace(string(data[:maxLen])) + "..."
	}
	return strings.TrimSpace(string(data))
}

// ---------------------------------------------------------------------------
// API endpoint methods
// ---------------------------------------------------------------------------

// Ping checks API server availability.
func (c *Client) Ping(ctx context.Context) (*VmmPingResponse, error) {
	var resp VmmPingResponse
	if _, err := c.doRequest(ctx, http.MethodGet, "/vmm.ping", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VMInfo returns general information about the VM instance.
func (c *Client) VMInfo(ctx context.Context) (*VmInfo, error) {
	var resp VmInfo
	if _, err := c.doRequest(ctx, http.MethodGet, "/vm.info", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateVM creates (but does not boot) the VM instance with the given
// configuration.
func (c *Client) CreateVM(ctx context.Context, cfg *VmConfig) error {
	_, err := c.doRequest(ctx, http.MethodPut, "/vm.create", cfg, nil)
	return err
}

// BootVM boots a previously created VM instance.
func (c *Client) BootVM(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodPut, "/vm.boot", nil, nil)
	return err
}

// ShutdownVM shuts down the VM instance.
func (c *Client) ShutdownVM(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodPut, "/vm.shutdown", nil, nil)
	return err
}

// DeleteVM deletes the VM instance (must be shut down first).
func (c *Client) DeleteVM(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodPut, "/vm.delete", nil, nil)
	return err
}

// AddDisk hotplugs a disk into the VM. Returns PciDeviceInfo on hotplug
// (HTTP 200 with body) and nil on cold plug (HTTP 204).
func (c *Client) AddDisk(ctx context.Context, cfg *DiskConfig) (*PciDeviceInfo, error) {
	var pci PciDeviceInfo
	decoded, err := c.doRequest(ctx, http.MethodPut, "/vm.add-disk", cfg, &pci)
	if err != nil {
		return nil, err
	}
	if !decoded {
		return nil, nil
	}
	return &pci, nil
}

// AddNet hotplugs a network device into the VM. Returns PciDeviceInfo on
// hotplug (HTTP 200 with body) and nil on cold plug (HTTP 204).
func (c *Client) AddNet(ctx context.Context, cfg *NetConfig) (*PciDeviceInfo, error) {
	var pci PciDeviceInfo
	decoded, err := c.doRequest(ctx, http.MethodPut, "/vm.add-net", cfg, &pci)
	if err != nil {
		return nil, err
	}
	if !decoded {
		return nil, nil
	}
	return &pci, nil
}

// RemoveDevice removes a device from the VM by its device ID.
func (c *Client) RemoveDevice(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, http.MethodPut, "/vm.remove-device", VmRemoveDevice{ID: id}, nil)
	return err
}

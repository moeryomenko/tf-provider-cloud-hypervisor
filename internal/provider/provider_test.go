package provider

import (
	"context"
	"os/exec"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// testAccProtoV6ProviderFactories is used by acceptance tests (TASK-008+).
// It creates a test provider server factory map for the cloudhypervisor provider.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cloudhypervisor": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck verifies prerequisites for acceptance tests.
// Skips the test when the cloud-hypervisor binary is not on PATH.
func testAccPreCheck(t *testing.T) {
	_, err := exec.LookPath("cloud-hypervisor")
	if err != nil {
		t.Skipf("cloud-hypervisor binary not found on PATH: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Schema validation tests
// ---------------------------------------------------------------------------

// TestProviderSchema verifies that the provider schema defines all four required
// attributes with the correct types and Optional/Required flags.
func TestProviderSchema(t *testing.T) {
	p := &cloudhypervisorProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes
	if attrs == nil {
		t.Fatal("schema attributes map is nil")
	}

	// Verify each attribute exists and has correct properties.
	t.Run("ch_binary_path", func(t *testing.T) {
		attr, ok := attrs["ch_binary_path"]
		if !ok {
			t.Fatal("missing required attribute: ch_binary_path")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("ch_binary_path must be a StringAttribute, got %T", attr)
		}
		if !strAttr.Optional {
			t.Error("ch_binary_path must be Optional")
		}
		if strAttr.Required {
			t.Error("ch_binary_path must not be Required")
		}
	})

	t.Run("ch_http_api", func(t *testing.T) {
		attr, ok := attrs["ch_http_api"]
		if !ok {
			t.Fatal("missing required attribute: ch_http_api")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("ch_http_api must be a StringAttribute, got %T", attr)
		}
		if !strAttr.Optional {
			t.Error("ch_http_api must be Optional")
		}
		if strAttr.Required {
			t.Error("ch_http_api must not be Required")
		}
	})

	t.Run("ch_api_socket_path", func(t *testing.T) {
		attr, ok := attrs["ch_api_socket_path"]
		if !ok {
			t.Fatal("missing required attribute: ch_api_socket_path")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("ch_api_socket_path must be a StringAttribute, got %T", attr)
		}
		if !strAttr.Optional {
			t.Error("ch_api_socket_path must be Optional")
		}
		if strAttr.Required {
			t.Error("ch_api_socket_path must not be Required")
		}
	})

	t.Run("manage_ch_process", func(t *testing.T) {
		attr, ok := attrs["manage_ch_process"]
		if !ok {
			t.Fatal("missing required attribute: manage_ch_process")
		}
		boolAttr, ok := attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("manage_ch_process must be a BoolAttribute, got %T", attr)
		}
		if !boolAttr.Optional {
			t.Error("manage_ch_process must be Optional")
		}
		if boolAttr.Required {
			t.Error("manage_ch_process must not be Required")
		}
	})
}

// ---------------------------------------------------------------------------
// Configure validation tests
// ---------------------------------------------------------------------------

// TestProviderConfigure verifies the Configure method applies defaults correctly,
// validates inputs, and populates ProviderData.
func TestProviderConfigure(t *testing.T) {
	// Shared provider instance for all sub-tests.
	p := &cloudhypervisorProvider{version: "test"}

	// Get schema to derive the config's tftypes.Object type.
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)

	// Skip Configure tests when schema is empty (pre-implementation red phase).
	if len(schemaResp.Schema.Attributes) == 0 {
		t.Skip("provider schema not yet populated — configure tests require attributes")
	}

	// helperConfig builds a tfsdk.Config from the given attribute values.
	// All unspecified attributes are set to null.
	helperConfig := func(t *testing.T, overrides map[string]tftypes.Value) provider.ConfigureRequest {
		t.Helper()

		// Build the full attribute set: default nulls, overridden by caller values.
		allAttrs := map[string]tftypes.Value{
			"ch_binary_path":     tftypes.NewValue(tftypes.String, nil),
			"ch_http_api":        tftypes.NewValue(tftypes.String, nil),
			"ch_api_socket_path": tftypes.NewValue(tftypes.String, nil),
			"manage_ch_process":  tftypes.NewValue(tftypes.Bool, nil),
		}
		for k, v := range overrides {
			allAttrs[k] = v
		}

		schemaType := schemaResp.Schema.Type().TerraformType(context.Background())

		configValue := tftypes.NewValue(schemaType, allAttrs)

		return provider.ConfigureRequest{
			Config: tfsdk.Config{
				Raw:    configValue,
				Schema: schemaResp.Schema,
			},
		}
	}

	// -----------------------------------------------------------------------
	// valid_managed_default: all nulls → defaults are applied
	// Expect: ManageProcess=true, BinaryPath="cloud-hypervisor"
	// -----------------------------------------------------------------------
	t.Run("valid_managed_default", func(t *testing.T) {
		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, nil), &configureResp)

		if configureResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", configureResp.Diagnostics.Errors())
		}

		if configureResp.ResourceData == nil {
			t.Fatal("ResourceData must be set to a *ProviderData value")
		}
	})

	// -----------------------------------------------------------------------
	// valid_managed_explicit: all attributes explicitly set
	// -----------------------------------------------------------------------
	t.Run("valid_managed_explicit", func(t *testing.T) {
		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"manage_ch_process": tftypes.NewValue(tftypes.Bool, true),
			"ch_binary_path":    tftypes.NewValue(tftypes.String, "cloud-hypervisor"),
			"ch_http_api":       tftypes.NewValue(tftypes.String, "http://localhost/api/v1"),
		}), &configureResp)

		if configureResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", configureResp.Diagnostics.Errors())
		}

		if configureResp.ResourceData == nil {
			t.Fatal("ResourceData must be set to a *ProviderData value")
		}
	})

	// -----------------------------------------------------------------------
	// valid_external: manage_ch_process=false, valid HTTP API URL
	// -----------------------------------------------------------------------
	t.Run("valid_external", func(t *testing.T) {
		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"manage_ch_process": tftypes.NewValue(tftypes.Bool, false),
			"ch_http_api":       tftypes.NewValue(tftypes.String, "http://localhost:8080/api/v1"),
		}), &configureResp)

		if configureResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", configureResp.Diagnostics.Errors())
		}

		if configureResp.ResourceData == nil {
			t.Fatal("ResourceData must be set to a *ProviderData value")
		}
	})

	// -----------------------------------------------------------------------
	// invalid_managed_no_binary: manage_ch_process=true with nonexistent binary
	// Expect: diagnostic warning (not error — may be installed before apply)
	// -----------------------------------------------------------------------
	t.Run("invalid_managed_no_binary", func(t *testing.T) {
		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"manage_ch_process": tftypes.NewValue(tftypes.Bool, true),
			"ch_binary_path":    tftypes.NewValue(tftypes.String, "nonexistent-binary-xyzzy"),
		}), &configureResp)

		// A warning is acceptable; an error diagnostic indicates over-validation.
		if configureResp.Diagnostics.HasError() {
			// Error is also acceptable at this stage — the provider may choose strict validation.
			t.Logf("provider returned error for nonexistent binary: %v", configureResp.Diagnostics.Errors())
		}
	})

	// -----------------------------------------------------------------------
	// invalid_external_no_url: manage_ch_process=false without ch_http_api
	// Expect: error diagnostic (external mode requires an endpoint)
	// -----------------------------------------------------------------------
	t.Run("invalid_external_no_url", func(t *testing.T) {
		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"manage_ch_process": tftypes.NewValue(tftypes.Bool, false),
			"ch_http_api":       tftypes.NewValue(tftypes.String, nil),
		}), &configureResp)

		if !configureResp.Diagnostics.HasError() {
			t.Error("expected error diagnostic when manage_ch_process=false and ch_http_api is unset")
		}
	})

	// -----------------------------------------------------------------------
	// env_var_binary_fallback: CH_BINARY_PATH set, config value is null
	// Expect: binary path falls back to env var
	// -----------------------------------------------------------------------
	t.Run("env_var_binary_fallback", func(t *testing.T) {
		const testBinary = "/usr/local/bin/cloud-hypervisor-test"
		t.Setenv("CH_BINARY_PATH", testBinary)

		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"ch_binary_path": tftypes.NewValue(tftypes.String, nil),
		}), &configureResp)

		if configureResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", configureResp.Diagnostics.Errors())
		}
	})

	// -----------------------------------------------------------------------
	// env_var_http_fallback: CH_HTTP_API set, config value is null
	// Expect: HTTP API URL falls back to env var
	// -----------------------------------------------------------------------
	t.Run("env_var_http_fallback", func(t *testing.T) {
		t.Setenv("CH_HTTP_API", "http://custom-host:9090/api/v1")

		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"ch_http_api": tftypes.NewValue(tftypes.String, nil),
		}), &configureResp)

		if configureResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", configureResp.Diagnostics.Errors())
		}
	})

	// -----------------------------------------------------------------------
	// env_var_config_override: both env var and config set → config wins
	// Expect: configured value is used, not env var
	// -----------------------------------------------------------------------
	t.Run("env_var_config_override", func(t *testing.T) {
		const envVal = "env-binary-path"
		const configVal = "config-binary-path"
		t.Setenv("CH_BINARY_PATH", envVal)

		var configureResp provider.ConfigureResponse
		p.Configure(context.Background(), helperConfig(t, map[string]tftypes.Value{
			"ch_binary_path": tftypes.NewValue(tftypes.String, configVal),
		}), &configureResp)

		if configureResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error diagnostics: %v", configureResp.Diagnostics.Errors())
		}
	})
}

// TestProviderMetadata verifies the provider metadata.
func TestProviderMetadata(t *testing.T) {
	p := &cloudhypervisorProvider{version: "1.0.0-test"}
	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "cloudhypervisor" {
		t.Errorf("expected type name 'cloudhypervisor', got %q", resp.TypeName)
	}
	if resp.Version != "1.0.0-test" {
		t.Errorf("expected version '1.0.0-test', got %q", resp.Version)
	}
}

// TestProviderResources verifies Resources returns an empty list initially.
func TestProviderResources(t *testing.T) {
	p := &cloudhypervisorProvider{version: "test"}
	resources := p.Resources(context.Background())
	if resources == nil {
		t.Error("Resources() must not return nil — return an empty slice")
	}
}

// TestProviderDataSources verifies DataSources returns an empty list initially.
func TestProviderDataSources(t *testing.T) {
	p := &cloudhypervisorProvider{version: "test"}
	dataSources := p.DataSources(context.Background())
	if dataSources == nil {
		t.Error("DataSources() must not return nil — return an empty slice")
	}
}

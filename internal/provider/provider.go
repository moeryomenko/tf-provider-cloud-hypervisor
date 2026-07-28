package provider

import (
	"context"
	"os"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/client"
)

// ProviderData carries the resolved provider configuration and a client
// factory. Configure creates this and attaches it to ResourceData/DataSourceData
// so resources can create their own per-VM CH API clients.
type ProviderData struct {
	BinaryPath    string
	HTTPAPI       string
	ManageProcess bool
	NewClient     func(endpoint string) (any, error)
}

type cloudhypervisorProvider struct {
	version string
}

// providerModel maps the provider-level schema attributes to Go types for
// config retrieval in Configure.
type providerModel struct {
	BinaryPath    types.String `tfsdk:"ch_binary_path"`
	HTTPAPI       types.String `tfsdk:"ch_http_api"`
	APISocketPath types.String `tfsdk:"ch_api_socket_path"`
	ManageProcess types.Bool   `tfsdk:"manage_ch_process"`
}

// New creates a new cloud-hypervisor provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cloudhypervisorProvider{version: version}
	}
}

func (p *cloudhypervisorProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cloudhypervisor"
	resp.Version = p.version
}

func (p *cloudhypervisorProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ch_binary_path": schema.StringAttribute{
				Optional: true,
			},
			"ch_http_api": schema.StringAttribute{
				Optional: true,
			},
			"ch_api_socket_path": schema.StringAttribute{
				Optional: true,
			},
			"manage_ch_process": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (p *cloudhypervisorProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model providerModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve manage_ch_process: default true.
	manageProcess := true
	if !model.ManageProcess.IsNull() {
		manageProcess = model.ManageProcess.ValueBool()
	}

	// Resolve binary path: config value → CH_BINARY_PATH env → default.
	binaryPath := "cloud-hypervisor"
	if !model.BinaryPath.IsNull() {
		binaryPath = model.BinaryPath.ValueString()
	} else if v, ok := os.LookupEnv("CH_BINARY_PATH"); ok {
		binaryPath = v
	}

	// Resolve HTTP API URL: config value → CH_HTTP_API env → default
	// (only applied in managed mode; external mode requires an explicit URL).
	httpAPI := ""
	if !model.HTTPAPI.IsNull() {
		httpAPI = model.HTTPAPI.ValueString()
	} else if v, ok := os.LookupEnv("CH_HTTP_API"); ok {
		httpAPI = v
	}

	if manageProcess {
		// Apply default URL for managed mode.
		if httpAPI == "" {
			httpAPI = "http://localhost/api/v1"
		}

		// Validate binary is resolvable at plan time. Emit a warning, not an
		// error — the binary may be installed before apply.
		if _, err := exec.LookPath(binaryPath); err != nil {
			resp.Diagnostics.AddWarning(
				"Cloud-Hypervisor Binary Not Found",
				"The cloud-hypervisor binary could not be found at "+binaryPath+". "+
					"It may be installed before resource creation. Error: "+err.Error(),
			)
		}
	} else {
		// External mode requires an explicit HTTP API URL.
		if httpAPI == "" {
			resp.Diagnostics.AddError(
				"External Mode Requires HTTP API URL",
				"When manage_ch_process is false, the provider cannot manage the "+
					"cloud-hypervisor process. You must set ch_http_api to point to "+
					"an existing cloud-hypervisor instance.",
			)
			return
		}
	}

	data := ProviderData{
		BinaryPath:    binaryPath,
		HTTPAPI:       httpAPI,
		ManageProcess: manageProcess,
		NewClient: func(endpoint string) (any, error) {
			if manageProcess {
				return client.New(endpoint)
			}
			return client.NewHTTP(endpoint)
		},
	}

	resp.ResourceData = &data
	resp.DataSourceData = &data
}

func (p *cloudhypervisorProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *cloudhypervisorProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

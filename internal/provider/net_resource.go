package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/client"
)

// Ensure the resource fully implements the Framework interfaces.
var (
	_ resource.Resource                = &netResource{}
	_ resource.ResourceWithConfigure   = &netResource{}
	_ resource.ResourceWithImportState = &netResource{}
)

// netResource implements resource.Resource for cloudhypervisor_net.
type netResource struct {
	providerData *ProviderData
}

// Configure receives provider-level data from the provider's Configure method.
func (r *netResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	pd, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Provider Data",
			fmt.Sprintf("Expected *ProviderData, got %T", req.ProviderData),
		)
		return
	}

	r.providerData = pd
}

// NewNetResource returns a new net resource for provider registration.
func NewNetResource() resource.Resource {
	return &netResource{}
}

// newNetResource returns a pointer to netResource for test access.
func newNetResource() *netResource {
	return &netResource{}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func (r *netResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "cloudhypervisor_net"
}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

type netResourceModel struct {
	VMSocketPath types.String `tfsdk:"vm_socket_path"`
	Tap          types.String `tfsdk:"tap"`
	IP           types.String `tfsdk:"ip"`
	Mask         types.String `tfsdk:"mask"`
	MAC          types.String `tfsdk:"mac"`
	HostMAC      types.String `tfsdk:"host_mac"`
	MTU          types.Int64  `tfsdk:"mtu"`
	Iommu        types.Bool   `tfsdk:"iommu"`
	NumQueues    types.Int64  `tfsdk:"num_queues"`
	QueueSize    types.Int64  `tfsdk:"queue_size"`
	VhostUser    types.Bool   `tfsdk:"vhost_user"`
	VhostSocket  types.String `tfsdk:"vhost_socket"`
	VhostMode    types.String `tfsdk:"vhost_mode"`
	ID           types.String `tfsdk:"id"`
	PCISegment   types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID  types.Int64  `tfsdk:"pci_device_id"`
	OffloadTSO   types.Bool   `tfsdk:"offload_tso"`
	OffloadUFO   types.Bool   `tfsdk:"offload_ufo"`
	OffloadCsum  types.Bool   `tfsdk:"offload_csum"`

	// Computed
	DeviceID types.String `tfsdk:"device_id"`
	BDF      types.String `tfsdk:"bdf"`
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func (r *netResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Hotplugs a network device into a running Cloud-Hypervisor VM.",
		Attributes: map[string]schema.Attribute{
			"vm_socket_path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the VM's API Unix domain socket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tap": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the host TAP interface",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Optional:    true,
				Description: "Virtual guest IP address",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mask": schema.StringAttribute{
				Optional:    true,
				Description: "Virtual guest network mask",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac": schema.StringAttribute{
				Optional:    true,
				Description: "MAC address for the virtual NIC",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_mac": schema.StringAttribute{
				Optional:    true,
				Description: "MAC address for the host TAP interface",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mtu": schema.Int64Attribute{
				Optional:    true,
				Description: "MTU for the virtual NIC",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"iommu": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable IOMMU for this device",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"num_queues": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of virtqueues",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"queue_size": schema.Int64Attribute{
				Optional:    true,
				Description: "Size of each virtqueue",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"vhost_user": schema.BoolAttribute{
				Optional:    true,
				Description: "Use vhost-user backend",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"vhost_socket": schema.StringAttribute{
				Optional:    true,
				Description: "vhost-user socket path",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vhost_mode": schema.StringAttribute{
				Optional:    true,
				Description: "vhost-user mode (client, server)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Device identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pci_segment": schema.Int64Attribute{
				Optional:    true,
				Description: "PCI segment group",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"pci_device_id": schema.Int64Attribute{
				Optional:    true,
				Description: "PCI device ID",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"offload_tso": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable TSO offload",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"offload_ufo": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable UFO offload",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"offload_csum": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable checksum offload",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			// Computed
			"device_id": schema.StringAttribute{
				Computed:    true,
				Description: "Device ID assigned by the CH API",
			},
			"bdf": schema.StringAttribute{
				Computed:    true,
				Description: "PCI BDF address assigned to the device",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Convert Terraform model to client NetConfig
// ---------------------------------------------------------------------------

func (m *netResourceModel) toClientConfig() *client.NetConfig {
	return &client.NetConfig{
		Tap:         m.Tap.ValueString(),
		IP:          m.IP.ValueString(),
		Mask:        m.Mask.ValueString(),
		MAC:         m.MAC.ValueString(),
		HostMAC:     m.HostMAC.ValueString(),
		MTU:         optionalInt(m.MTU),
		Iommu:       optionalBool(m.Iommu),
		NumQueues:   optionalInt(m.NumQueues),
		QueueSize:   optionalInt(m.QueueSize),
		VhostUser:   optionalBool(m.VhostUser),
		VhostSocket: m.VhostSocket.ValueString(),
		VhostMode:   m.VhostMode.ValueString(),
		ID:          m.ID.ValueString(),
		PCISegment:  optionalInt16(m.PCISegment),
		PCIDeviceID: optionalUint8(m.PCIDeviceID),
		OffloadTSO:  optionalBool(m.OffloadTSO),
		OffloadUFO:  optionalBool(m.OffloadUFO),
		OffloadCsum: optionalBool(m.OffloadCsum),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (r *netResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan netResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	socketPath := plan.VMSocketPath.ValueString()
	ch, err := client.New(socketPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create API Client",
			"Could not connect to VM socket: "+err.Error(),
		)
		return
	}

	cfg := plan.toClientConfig()
	pciInfo, err := ch.AddNet(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Hotplug Net",
			"The network device could not be added to the VM. Error: "+err.Error(),
		)
		return
	}

	if pciInfo != nil {
		plan.DeviceID = types.StringValue(pciInfo.ID)
		plan.BDF = types.StringValue(pciInfo.BDF)
	} else {
		plan.DeviceID = types.StringValue(plan.ID.ValueString())
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func (r *netResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state netResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	socketPath := state.VMSocketPath.ValueString()
	ch, err := client.New(socketPath)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to Create API Client",
			"Could not connect to VM socket for Read: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	_, err = ch.VMInfo(ctx)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddWarning(
			"Failed to Read VM Info",
			"Could not verify VM state: "+err.Error(),
		)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func (r *netResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state netResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := state.DeviceID.ValueString()
	if deviceID == "" {
		deviceID = state.ID.ValueString()
	}
	if deviceID == "" {
		return
	}

	socketPath := state.VMSocketPath.ValueString()
	ch, err := client.New(socketPath)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to Create API Client",
			"Could not connect to VM socket for Delete: "+err.Error(),
		)
		return
	}

	if err := ch.RemoveDevice(ctx, deviceID); err != nil {
		if !errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Failed to Remove Net Device",
				"The network device could not be removed. Error: "+err.Error(),
			)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Update (not supported — all RequiresReplace)
// ---------------------------------------------------------------------------

func (r *netResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The cloudhypervisor_net resource does not support in-place updates. "+
			"Any change to the configuration forces resource recreation.",
	)
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func (r *netResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ",", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected format: <socket_path>,<device_id>",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vm_socket_path"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), parts[1])...)
}

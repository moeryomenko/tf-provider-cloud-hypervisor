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
	_ resource.Resource                = &diskResource{}
	_ resource.ResourceWithConfigure   = &diskResource{}
	_ resource.ResourceWithImportState = &diskResource{}
)

// diskResource implements resource.Resource for cloudhypervisor_disk.
type diskResource struct {
	providerData *ProviderData
}

// Configure receives provider-level data from the provider's Configure method.
func (r *diskResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// NewDiskResource returns a new disk resource for provider registration.
func NewDiskResource() resource.Resource {
	return &diskResource{}
}

// newDiskResource returns a pointer to diskResource for test access.
func newDiskResource() *diskResource {
	return &diskResource{}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func (r *diskResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "cloudhypervisor_disk"
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

type diskResourceModel struct {
	VMSocketPath    types.String `tfsdk:"vm_socket_path"`
	Path            types.String `tfsdk:"path"`
	Readonly        types.Bool   `tfsdk:"readonly"`
	Direct          types.Bool   `tfsdk:"direct"`
	Iommu           types.Bool   `tfsdk:"iommu"`
	NumQueues       types.Int64  `tfsdk:"num_queues"`
	QueueSize       types.Int64  `tfsdk:"queue_size"`
	VhostUser       types.Bool   `tfsdk:"vhost_user"`
	VhostSocket     types.String `tfsdk:"vhost_socket"`
	PCISegment      types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID     types.Int64  `tfsdk:"pci_device_id"`
	ID              types.String `tfsdk:"id"`
	Serial          types.String `tfsdk:"serial"`
	RateLimitGroup  types.String `tfsdk:"rate_limit_group"`
	BackingFiles    types.Bool   `tfsdk:"backing_files"`
	Sparse          types.Bool   `tfsdk:"sparse"`
	ImageType       types.String `tfsdk:"image_type"`
	LockGranularity types.String `tfsdk:"lock_granularity"`

	// Computed
	DeviceID types.String `tfsdk:"device_id"`
	BDF      types.String `tfsdk:"bdf"`
}

func (r *diskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Hotplugs a disk device into a running Cloud-Hypervisor VM.",
		Attributes: map[string]schema.Attribute{
			"vm_socket_path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the VM's API Unix domain socket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the disk image",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"readonly": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the disk is read-only",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"direct": schema.BoolAttribute{
				Optional:    true,
				Description: "Use direct I/O for the disk",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
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
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Device identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"serial": schema.StringAttribute{
				Optional:    true,
				Description: "Disk serial number",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rate_limit_group": schema.StringAttribute{
				Optional:    true,
				Description: "Rate limit group ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"backing_files": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable backing files support",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"sparse": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable sparse file support",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"image_type": schema.StringAttribute{
				Optional:    true,
				Description: "Disk image type (raw, qcow2, etc.)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lock_granularity": schema.StringAttribute{
				Optional:    true,
				Description: "Lock granularity for disk access",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
// Convert Terraform model to client DiskConfig
// ---------------------------------------------------------------------------

func (m *diskResourceModel) toClientConfig() *client.DiskConfig {
	return &client.DiskConfig{
		Path:            m.Path.ValueString(),
		Readonly:        optionalBool(m.Readonly),
		Direct:          optionalBool(m.Direct),
		Iommu:           optionalBool(m.Iommu),
		NumQueues:       optionalInt(m.NumQueues),
		QueueSize:       optionalInt(m.QueueSize),
		VhostUser:       optionalBool(m.VhostUser),
		VhostSocket:     m.VhostSocket.ValueString(),
		PCISegment:      optionalInt16(m.PCISegment),
		PCIDeviceID:     optionalUint8(m.PCIDeviceID),
		ID:              m.ID.ValueString(),
		Serial:          m.Serial.ValueString(),
		RateLimitGroup:  m.RateLimitGroup.ValueString(),
		BackingFiles:    optionalBool(m.BackingFiles),
		Sparse:          optionalBool(m.Sparse),
		ImageType:       optionalImageType(m.ImageType),
		LockGranularity: optionalLockGranularity(m.LockGranularity),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (r *diskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan diskResourceModel
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
	pciInfo, err := ch.AddDisk(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Hotplug Disk",
			"The disk could not be added to the VM. Error: "+err.Error(),
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

func (r *diskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state diskResourceModel
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

	// Use VM info to verify the VM still exists.
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

func (r *diskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state diskResourceModel
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
		// If the VM no longer exists, the device is implicitly removed.
		if !errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Failed to Remove Disk",
				"The disk could not be removed. Error: "+err.Error(),
			)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Update (not supported — all RequiresReplace)
// ---------------------------------------------------------------------------

func (r *diskResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The cloudhypervisor_disk resource does not support in-place updates. "+
			"Any change to the configuration forces resource recreation.",
	)
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func (r *diskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import expects comma-separated: socket_path,device_id
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

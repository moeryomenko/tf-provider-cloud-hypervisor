package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/chproc"
	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/client"
)

// Ensure the resource fully implements the Framework interfaces.
var (
	_ resource.Resource                = &vmResource{}
	_ resource.ResourceWithConfigure   = &vmResource{}
	_ resource.ResourceWithImportState = &vmResource{}
)

// vmResource implements resource.Resource for cloudhypervisor_vm.
type vmResource struct {
	// providerData is set by Configure and contains the resolved provider
	// configuration (binary path, HTTP API URL, manage-process mode) plus
	// a client factory.
	providerData *ProviderData
}

// Configure receives provider-level data from the provider's Configure
// method. It caches the *ProviderData on the struct for use in CRUD
// methods.
func (r *vmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// NewVMResource returns a new VM resource for provider registration.
func NewVMResource() resource.Resource {
	return &vmResource{}
}

// newVMResource returns a pointer to vmResource for test access.
func newVMResource() *vmResource {
	return &vmResource{}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func (r *vmResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "cloudhypervisor_vm"
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

func (r *vmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cloud-Hypervisor VM instance. " +
			"Creates the VM config, boots it, and manages the underlying cloud-hypervisor process.",
		Attributes: map[string]schema.Attribute{
			// Computed attributes for process management.
			"socket_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path to the VM's API Unix domain socket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"socket_dir": schema.StringAttribute{
				Computed:    true,
				Description: "Directory containing the VM's API socket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// VmConfig field groups.
			"payload": schema.SingleNestedAttribute{
				Required:    true,
				Description: "VM payload configuration (kernel, initramfs, cmdline)",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"firmware":  schema.StringAttribute{Optional: true, Description: "Path to firmware binary"},
					"kernel":    schema.StringAttribute{Optional: true, Description: "Path to kernel binary"},
					"cmdline":   schema.StringAttribute{Optional: true, Description: "Kernel command line"},
					"initramfs": schema.StringAttribute{Optional: true, Description: "Path to initramfs image"},
					"igvm":      schema.StringAttribute{Optional: true, Description: "Path to IGVM file"},
					"host_data": schema.StringAttribute{Optional: true, Description: "Host data blob"},
				},
			},
			"cpus": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "CPU configuration (topology, features, affinity)",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"boot_vcpus":      schema.Int64Attribute{Required: true, Description: "Number of boot vCPUs"},
					"max_vcpus":       schema.Int64Attribute{Required: true, Description: "Maximum number of vCPUs"},
					"kvm_hyperv":      schema.BoolAttribute{Optional: true, Description: "Enable KVM hyperv extensions"},
					"max_phys_bits":   schema.Int64Attribute{Optional: true, Description: "Maximum physical address bits"},
					"nested":          schema.BoolAttribute{Optional: true, Description: "Enable nested virtualization"},
					"core_scheduling": schema.StringAttribute{Optional: true, Description: "Core scheduling mode"},
					"topology": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"threads_per_core": schema.Int64Attribute{Optional: true},
							"cores_per_die":    schema.Int64Attribute{Optional: true},
							"dies_per_package": schema.Int64Attribute{Optional: true},
							"packages":         schema.Int64Attribute{Optional: true},
						},
					},
					"affinity": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"vcpu":      schema.Int64Attribute{Required: true},
								"host_cpus": schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
							},
						},
					},
					"features": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"amx": schema.BoolAttribute{Optional: true},
						},
					},
				},
			},
			"memory": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Memory configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"size":            schema.Int64Attribute{Required: true, Description: "Memory size in bytes"},
					"hotplug_size":    schema.Int64Attribute{Optional: true},
					"hotplugged_size": schema.Int64Attribute{Optional: true},
					"mergeable":       schema.BoolAttribute{Optional: true},
					"hotplug_method":  schema.StringAttribute{Optional: true},
					"shared":          schema.BoolAttribute{Optional: true},
					"hugepages":       schema.BoolAttribute{Optional: true},
					"hugepage_size":   schema.Int64Attribute{Optional: true},
					"prefault":        schema.BoolAttribute{Optional: true},
					"reserve":         schema.BoolAttribute{Optional: true},
					"thp":             schema.BoolAttribute{Optional: true},
					"zones": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id":              schema.StringAttribute{Required: true},
								"size":            schema.Int64Attribute{Required: true},
								"file":            schema.StringAttribute{Optional: true},
								"mergeable":       schema.BoolAttribute{Optional: true},
								"shared":          schema.BoolAttribute{Optional: true},
								"hugepages":       schema.BoolAttribute{Optional: true},
								"hugepage_size":   schema.Int64Attribute{Optional: true},
								"host_numa_node":  schema.Int64Attribute{Optional: true},
								"hotplug_size":    schema.Int64Attribute{Optional: true},
								"hotplugged_size": schema.Int64Attribute{Optional: true},
								"prefault":        schema.BoolAttribute{Optional: true},
								"reserve":         schema.BoolAttribute{Optional: true},
							},
						},
					},
				},
			},
			"rate_limit_groups": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Rate limit groups",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{Required: true},
						"rate_limiter": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"bandwidth": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"size":           schema.Int64Attribute{Required: true},
										"one_time_burst": schema.Int64Attribute{Optional: true},
										"refill_time":    schema.Int64Attribute{Required: true},
									},
								},
								"ops": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"size":           schema.Int64Attribute{Required: true},
										"one_time_burst": schema.Int64Attribute{Optional: true},
										"refill_time":    schema.Int64Attribute{Required: true},
									},
								},
							},
						},
					},
				},
			},
			"disks": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Create-time disk devices (use cloudhypervisor_disk for hotplug)",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path":             schema.StringAttribute{Optional: true},
						"readonly":         schema.BoolAttribute{Optional: true},
						"direct":           schema.BoolAttribute{Optional: true},
						"iommu":            schema.BoolAttribute{Optional: true},
						"num_queues":       schema.Int64Attribute{Optional: true},
						"queue_size":       schema.Int64Attribute{Optional: true},
						"vhost_user":       schema.BoolAttribute{Optional: true},
						"vhost_socket":     schema.StringAttribute{Optional: true},
						"pci_segment":      schema.Int64Attribute{Optional: true},
						"pci_device_id":    schema.Int64Attribute{Optional: true},
						"id":               schema.StringAttribute{Optional: true},
						"serial":           schema.StringAttribute{Optional: true},
						"rate_limit_group": schema.StringAttribute{Optional: true},
						"backing_files":    schema.BoolAttribute{Optional: true},
						"sparse":           schema.BoolAttribute{Optional: true},
						"image_type":       schema.StringAttribute{Optional: true},
						"lock_granularity": schema.StringAttribute{Optional: true},
						"rate_limiter": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"bandwidth": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"size":           schema.Int64Attribute{Required: true},
										"one_time_burst": schema.Int64Attribute{Optional: true},
										"refill_time":    schema.Int64Attribute{Required: true},
									},
								},
								"ops": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"size":           schema.Int64Attribute{Required: true},
										"one_time_burst": schema.Int64Attribute{Optional: true},
										"refill_time":    schema.Int64Attribute{Required: true},
									},
								},
							},
						},
					},
				},
			},
			"net": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Create-time network devices (use cloudhypervisor_net for hotplug)",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tap":           schema.StringAttribute{Optional: true},
						"ip":            schema.StringAttribute{Optional: true},
						"mask":          schema.StringAttribute{Optional: true},
						"mac":           schema.StringAttribute{Optional: true},
						"host_mac":      schema.StringAttribute{Optional: true},
						"mtu":           schema.Int64Attribute{Optional: true},
						"iommu":         schema.BoolAttribute{Optional: true},
						"num_queues":    schema.Int64Attribute{Optional: true},
						"queue_size":    schema.Int64Attribute{Optional: true},
						"vhost_user":    schema.BoolAttribute{Optional: true},
						"vhost_socket":  schema.StringAttribute{Optional: true},
						"vhost_mode":    schema.StringAttribute{Optional: true},
						"id":            schema.StringAttribute{Optional: true},
						"pci_segment":   schema.Int64Attribute{Optional: true},
						"pci_device_id": schema.Int64Attribute{Optional: true},
						"offload_tso":   schema.BoolAttribute{Optional: true},
						"offload_ufo":   schema.BoolAttribute{Optional: true},
						"offload_csum":  schema.BoolAttribute{Optional: true},
						"rate_limiter": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"bandwidth": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"size":           schema.Int64Attribute{Required: true},
										"one_time_burst": schema.Int64Attribute{Optional: true},
										"refill_time":    schema.Int64Attribute{Required: true},
									},
								},
								"ops": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"size":           schema.Int64Attribute{Required: true},
										"one_time_burst": schema.Int64Attribute{Optional: true},
										"refill_time":    schema.Int64Attribute{Required: true},
									},
								},
							},
						},
					},
				},
			},
			"rng": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Random number generator device",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"src":           schema.StringAttribute{Required: true},
					"id":            schema.StringAttribute{Optional: true},
					"pci_segment":   schema.Int64Attribute{Optional: true},
					"pci_device_id": schema.Int64Attribute{Optional: true},
					"iommu":         schema.BoolAttribute{Optional: true},
				},
			},
			"balloon": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Virtio-balloon device",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"size":                schema.Int64Attribute{Required: true},
					"id":                  schema.StringAttribute{Optional: true},
					"pci_segment":         schema.Int64Attribute{Optional: true},
					"pci_device_id":       schema.Int64Attribute{Optional: true},
					"iommu":               schema.BoolAttribute{Optional: true},
					"deflate_on_oom":      schema.BoolAttribute{Optional: true},
					"free_page_reporting": schema.BoolAttribute{Optional: true},
				},
			},
			"fs": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Virtio-fs shared directories",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tag":           schema.StringAttribute{Required: true},
						"socket":        schema.StringAttribute{Required: true},
						"num_queues":    schema.Int64Attribute{Optional: true},
						"queue_size":    schema.Int64Attribute{Optional: true},
						"pci_segment":   schema.Int64Attribute{Optional: true},
						"pci_device_id": schema.Int64Attribute{Optional: true},
						"id":            schema.StringAttribute{Optional: true},
					},
				},
			},
			"generic_vhost_user": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Generic vhost-user devices",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"socket":        schema.StringAttribute{Required: true},
						"queue_sizes":   schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
						"device_type":   schema.Int64Attribute{Required: true},
						"pci_segment":   schema.Int64Attribute{Optional: true},
						"pci_device_id": schema.Int64Attribute{Optional: true},
					},
				},
			},
			"pmem": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Persistent memory devices",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"file":           schema.StringAttribute{Required: true},
						"size":           schema.Int64Attribute{Optional: true},
						"iommu":          schema.BoolAttribute{Optional: true},
						"discard_writes": schema.BoolAttribute{Optional: true},
						"pci_segment":    schema.Int64Attribute{Optional: true},
						"pci_device_id":  schema.Int64Attribute{Optional: true},
						"id":             schema.StringAttribute{Optional: true},
					},
				},
			},
			"serial": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Serial port configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"file":   schema.StringAttribute{Optional: true},
					"socket": schema.StringAttribute{Optional: true},
					"mode":   schema.StringAttribute{Required: true},
				},
			},
			"console": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Console device configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"file":          schema.StringAttribute{Optional: true},
					"socket":        schema.StringAttribute{Optional: true},
					"mode":          schema.StringAttribute{Required: true},
					"iommu":         schema.BoolAttribute{Optional: true},
					"id":            schema.StringAttribute{Optional: true},
					"pci_segment":   schema.Int64Attribute{Optional: true},
					"pci_device_id": schema.Int64Attribute{Optional: true},
				},
			},
			"debug_console": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Debug console configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"file":   schema.StringAttribute{Optional: true},
					"mode":   schema.StringAttribute{Required: true},
					"iobase": schema.Int64Attribute{Optional: true},
				},
			},
			"devices": schema.ListNestedAttribute{
				Optional:    true,
				Description: "VFIO device passthrough",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path":                  schema.StringAttribute{Optional: true},
						"iommu":                 schema.BoolAttribute{Optional: true},
						"pci_segment":           schema.Int64Attribute{Optional: true},
						"pci_device_id":         schema.Int64Attribute{Optional: true},
						"id":                    schema.StringAttribute{Optional: true},
						"x_nv_gpudirect_clique": schema.Int64Attribute{Optional: true},
					},
				},
			},
			"user_devices": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Userspace (vhost-user) devices",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"socket":        schema.StringAttribute{Required: true},
						"id":            schema.StringAttribute{Optional: true},
						"pci_segment":   schema.Int64Attribute{Optional: true},
						"pci_device_id": schema.Int64Attribute{Optional: true},
					},
				},
			},
			"vdpa": schema.ListNestedAttribute{
				Optional:    true,
				Description: "vDPA devices",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path":          schema.StringAttribute{Required: true},
						"num_queues":    schema.Int64Attribute{Required: true},
						"iommu":         schema.BoolAttribute{Optional: true},
						"pci_segment":   schema.Int64Attribute{Optional: true},
						"pci_device_id": schema.Int64Attribute{Optional: true},
						"id":            schema.StringAttribute{Optional: true},
					},
				},
			},
			"vsock": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Vhost-user socket (vsock) device",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"cid":           schema.Int64Attribute{Required: true},
					"socket":        schema.StringAttribute{Required: true},
					"iommu":         schema.BoolAttribute{Optional: true},
					"pci_segment":   schema.Int64Attribute{Optional: true},
					"pci_device_id": schema.Int64Attribute{Optional: true},
					"id":            schema.StringAttribute{Optional: true},
				},
			},
			"numa": schema.ListNestedAttribute{
				Optional:    true,
				Description: "NUMA node configuration",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"guest_numa_id": schema.Int64Attribute{Required: true},
						"cpus":          schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
						"memory_zones":  schema.ListAttribute{Optional: true, ElementType: types.StringType},
						"pci_segments":  schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
						"device_id":     schema.StringAttribute{Optional: true},
						"distances": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"destination": schema.Int64Attribute{Required: true},
									"distance":    schema.Int64Attribute{Required: true},
								},
							},
						},
					},
				},
			},
			"iommu": schema.BoolAttribute{
				Optional:    true,
				Description: "Global IOMMU flag",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"watchdog": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable watchdog device",
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"rtc": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "RTC device configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"id":            schema.StringAttribute{Optional: true},
					"pci_segment":   schema.Int64Attribute{Optional: true},
					"pci_device_id": schema.Int64Attribute{Optional: true},
					"iommu":         schema.BoolAttribute{Optional: true},
				},
			},
			"pvpanic": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable pvpanic device",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"pci_segments": schema.ListNestedAttribute{
				Optional:    true,
				Description: "PCI segment configuration",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pci_segment":            schema.Int64Attribute{Required: true},
						"mmio32_aperture_weight": schema.Int64Attribute{Optional: true},
						"mmio64_aperture_weight": schema.Int64Attribute{Optional: true},
					},
				},
			},
			"platform": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Platform configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"num_pci_segments":         schema.Int64Attribute{Optional: true},
					"iommu_segments":           schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
					"iommu_address_width_bits": schema.Int64Attribute{Optional: true},
					"system_serial_number":     schema.StringAttribute{Optional: true},
					"serial_number":            schema.StringAttribute{Optional: true},
					"system_uuid":              schema.StringAttribute{Optional: true},
					"uuid":                     schema.StringAttribute{Optional: true},
					"oem_strings":              schema.ListAttribute{Optional: true, ElementType: types.StringType},
					"system_manufacturer":      schema.StringAttribute{Optional: true},
					"system_product_name":      schema.StringAttribute{Optional: true},
					"system_version":           schema.StringAttribute{Optional: true},
					"system_family":            schema.StringAttribute{Optional: true},
					"system_sku_number":        schema.StringAttribute{Optional: true},
					"chassis_asset_tag":        schema.StringAttribute{Optional: true},
					"tdx":                      schema.BoolAttribute{Optional: true},
					"sev_snp":                  schema.BoolAttribute{Optional: true},
					"iommufd":                  schema.BoolAttribute{Optional: true},
					"vfio_p2p_dma":             schema.BoolAttribute{Optional: true},
				},
			},
			"tpm": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "TPM device configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"socket": schema.StringAttribute{Required: true},
				},
			},
			"landlock_enable": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable Landlock sandboxing",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"landlock_rules": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Landlock access rules",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path":   schema.StringAttribute{Required: true},
						"access": schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (r *vmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform plan to CH API config.
	cfg := plan.toClientConfig()

	// Resolve socket path.
	pd := r.providerData
	socketPath := pd.HTTPAPI
	var mgr *chproc.Manager

	if pd.ManageProcess {
		// Start the cloud-hypervisor process.
		mgr = chproc.NewManager(chproc.WithBinaryPath(pd.BinaryPath))
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		var err error
		socketPath, err = mgr.Start(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Start Cloud-Hypervisor",
				"The cloud-hypervisor process could not be started. Error: "+err.Error(),
			)
			return
		}

		// Wait for the API socket to become reachable.
		if err := mgr.WaitReady(ctx, 30*time.Second); err != nil {
			_ = mgr.Stop(context.Background())
			resp.Diagnostics.AddError(
				"Cloud-Hypervisor Socket Not Ready",
				"The API socket did not become reachable. Error: "+err.Error(),
			)
			return
		}

		plan.SocketDir = types.StringValue(filepath.Dir(socketPath))
	} else {
		// External mode: use the configured HTTP API URL.
		mgr = nil
	}

	plan.SocketPath = types.StringValue(socketPath)

	// Create an API client for the VM's socket.
	ch, err := client.New(socketPath)
	if err != nil {
		if mgr != nil {
			_ = mgr.Stop(context.Background())
		}
		resp.Diagnostics.AddError(
			"Failed to Create API Client",
			"Could not create CH API client: "+err.Error(),
		)
		return
	}

	// Create the VM configuration.
	if err := ch.CreateVM(ctx, cfg); err != nil {
		if mgr != nil {
			_ = mgr.Stop(context.Background())
		}
		resp.Diagnostics.AddError(
			"Failed to Create VM",
			"The VM could not be created. Error: "+err.Error(),
		)
		return
	}

	// Boot the VM.
	if err := ch.BootVM(ctx); err != nil {
		// Attempt cleanup on boot failure.
		_ = ch.ShutdownVM(context.Background())
		_ = ch.DeleteVM(context.Background())
		if mgr != nil {
			_ = mgr.Stop(context.Background())
		}
		resp.Diagnostics.AddError(
			"Failed to Boot VM",
			"The VM could not be booted. Error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// ---------------------------------------------------------------------------
// Read
// ---------------------------------------------------------------------------

func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	socketPath := state.SocketPath.ValueString()
	if socketPath == "" {
		// Resource has no socket path — it was never created or is in a bad state.
		resp.State.RemoveResource(ctx)
		return
	}

	ch, err := client.New(socketPath)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to Create API Client",
			"Could not create CH API client for Read: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	info, err := ch.VMInfo(ctx)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			// VM has been destroyed outside Terraform.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Read VM",
			"Could not get VM info: "+err.Error(),
		)
		return
	}

	// Update state from live VM info.
	_ = info // VmInfo contains config+state but we preserve our plan config
	// We keep the state as-is since we store the config in state,
	// and the config from VmInfo may have defaults applied server-side.

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	socketPath := state.SocketPath.ValueString()
	if socketPath == "" {
		return
	}

	ch, err := client.New(socketPath)
	if err == nil {
		// Attempt graceful shutdown.
		_ = ch.ShutdownVM(ctx)
		_ = ch.DeleteVM(ctx)
	}

	// In managed mode, kill the cloud-hypervisor process.
	// The process PID is tracked via the socket directory.
	if r.providerData != nil && r.providerData.ManageProcess && !state.SocketDir.IsNull() {
		socketDir := state.SocketDir.ValueString()
		if socketDir != "" {
			_ = killProcessBySocketDir(socketDir)
		}
	}
}

// killProcessBySocketDir finds a cloud-hypervisor process whose
// --api-socket-path matches the given directory and kills it.
func killProcessBySocketDir(socketDir string) error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("read /proc: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pidStr := entry.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		cmdline, err := os.ReadFile(filepath.Join("/proc", pidStr, "cmdline"))
		if err != nil {
			continue
		}

		cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
		if !strings.Contains(cmdStr, socketDir) {
			continue
		}

		// Signal graceful stop.
		_ = syscall.Kill(pid, syscall.SIGTERM)

		// Wait up to 5 seconds for process to exit.
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if err := syscall.Kill(pid, 0); err != nil {
				_, _ = syscall.Wait4(pid, nil, 0, nil)
				_ = os.RemoveAll(socketDir)
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}

		// Force kill.
		_ = syscall.Kill(pid, syscall.SIGKILL)
		_, _ = syscall.Wait4(pid, nil, 0, nil)
		_ = os.RemoveAll(socketDir)
		return nil
	}

	return nil
}

// ---------------------------------------------------------------------------
// Update (not supported — all RequiresReplace)
// ---------------------------------------------------------------------------

func (r *vmResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The cloudhypervisor_vm resource does not support in-place updates. "+
			"Any change to the configuration forces resource recreation.",
	)
}

// ---------------------------------------------------------------------------
// ImportState
// ---------------------------------------------------------------------------

func (r *vmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by socket path: the ID should be the path to the API socket
	// (e.g., /tmp/ch-tf-xxxxx/api.sock).
	resource.ImportStatePassthroughID(ctx, path.Root("socket_path"), req, resp)
}

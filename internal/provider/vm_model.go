package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/client"
)

// ---------------------------------------------------------------------------
// vmResourceModel — Terraform state for cloudhypervisor_vm
// ---------------------------------------------------------------------------

// vmResourceModel maps the full VmConfig to Terraform state attributes.
// Every field group from the CH API VmConfig is represented as a nested
// attribute. Computed fields (socket_path, socket_dir) are set by the
// provider during Create.
type vmResourceModel struct {
	// Provider-managed process attributes (computed).
	SocketPath types.String `tfsdk:"socket_path"`
	SocketDir  types.String `tfsdk:"socket_dir"`

	// VmConfig field groups — all from client.VmConfig.
	Payload          *vmPayloadModel           `tfsdk:"payload"`
	Cpus             *vmCpusModel              `tfsdk:"cpus"`
	Memory           *vmMemoryModel            `tfsdk:"memory"`
	RateLimitGroups  []vmRateLimitGroupModel   `tfsdk:"rate_limit_groups"`
	Disks            []vmDiskModel             `tfsdk:"disks"`
	Net              []vmNetModel              `tfsdk:"net"`
	Rng              *vmRngModel               `tfsdk:"rng"`
	Balloon          *vmBalloonModel           `tfsdk:"balloon"`
	FS               []vmFsModel               `tfsdk:"fs"`
	GenericVhostUser []vmGenericVhostUserModel `tfsdk:"generic_vhost_user"`
	Pmem             []vmPmemModel             `tfsdk:"pmem"`
	Serial           *vmSerialModel            `tfsdk:"serial"`
	Console          *vmConsoleModel           `tfsdk:"console"`
	DebugConsole     *vmDebugConsoleModel      `tfsdk:"debug_console"`
	Devices          []vmDeviceModel           `tfsdk:"devices"`
	UserDevices      []vmUserDeviceModel       `tfsdk:"user_devices"`
	Vdpa             []vmVdpaModel             `tfsdk:"vdpa"`
	Vsock            *vmVsockModel             `tfsdk:"vsock"`
	Numa             []vmNumaModel             `tfsdk:"numa"`
	Iommu            types.Bool                `tfsdk:"iommu"`
	Watchdog         types.Bool                `tfsdk:"watchdog"`
	Rtc              *vmRtcModel               `tfsdk:"rtc"`
	Pvpanic          types.Bool                `tfsdk:"pvpanic"`
	PCISegments      []vmPciSegmentModel       `tfsdk:"pci_segments"`
	Platform         *vmPlatformModel          `tfsdk:"platform"`
	Tpm              *vmTpmModel               `tfsdk:"tpm"`
	LandlockEnable   types.Bool                `tfsdk:"landlock_enable"`
	LandlockRules    []vmLandlockRuleModel     `tfsdk:"landlock_rules"`
}

// ---------------------------------------------------------------------------
// Payload
// ---------------------------------------------------------------------------

type vmPayloadModel struct {
	Firmware  types.String `tfsdk:"firmware"`
	Kernel    types.String `tfsdk:"kernel"`
	Cmdline   types.String `tfsdk:"cmdline"`
	Initramfs types.String `tfsdk:"initramfs"`
	IGVM      types.String `tfsdk:"igvm"`
	HostData  types.String `tfsdk:"host_data"`
}

// ---------------------------------------------------------------------------
// CPU
// ---------------------------------------------------------------------------

type vmCpusModel struct {
	BootVcpus      types.Int64          `tfsdk:"boot_vcpus"`
	MaxVcpus       types.Int64          `tfsdk:"max_vcpus"`
	Topology       *vmCpuTopologyModel  `tfsdk:"topology"`
	KvmHyperv      types.Bool           `tfsdk:"kvm_hyperv"`
	MaxPhysBits    types.Int64          `tfsdk:"max_phys_bits"`
	Nested         types.Bool           `tfsdk:"nested"`
	Affinity       []vmCpuAffinityModel `tfsdk:"affinity"`
	Features       *vmCpuFeaturesModel  `tfsdk:"features"`
	CoreScheduling types.String         `tfsdk:"core_scheduling"`
}

type vmCpuTopologyModel struct {
	ThreadsPerCore types.Int64 `tfsdk:"threads_per_core"`
	CoresPerDie    types.Int64 `tfsdk:"cores_per_die"`
	DiesPerPackage types.Int64 `tfsdk:"dies_per_package"`
	Packages       types.Int64 `tfsdk:"packages"`
}

type vmCpuAffinityModel struct {
	Vcpu     types.Int64   `tfsdk:"vcpu"`
	HostCPUs []types.Int64 `tfsdk:"host_cpus"`
}

type vmCpuFeaturesModel struct {
	Amx types.Bool `tfsdk:"amx"`
}

// ---------------------------------------------------------------------------
// Memory
// ---------------------------------------------------------------------------

type vmMemoryModel struct {
	Size           types.Int64         `tfsdk:"size"`
	HotplugSize    types.Int64         `tfsdk:"hotplug_size"`
	HotpluggedSize types.Int64         `tfsdk:"hotplugged_size"`
	Mergeable      types.Bool          `tfsdk:"mergeable"`
	HotplugMethod  types.String        `tfsdk:"hotplug_method"`
	Shared         types.Bool          `tfsdk:"shared"`
	Hugepages      types.Bool          `tfsdk:"hugepages"`
	HugepageSize   types.Int64         `tfsdk:"hugepage_size"`
	Prefault       types.Bool          `tfsdk:"prefault"`
	Reserve        types.Bool          `tfsdk:"reserve"`
	Thp            types.Bool          `tfsdk:"thp"`
	Zones          []vmMemoryZoneModel `tfsdk:"zones"`
}

type vmMemoryZoneModel struct {
	ID             types.String `tfsdk:"id"`
	Size           types.Int64  `tfsdk:"size"`
	File           types.String `tfsdk:"file"`
	Mergeable      types.Bool   `tfsdk:"mergeable"`
	Shared         types.Bool   `tfsdk:"shared"`
	Hugepages      types.Bool   `tfsdk:"hugepages"`
	HugepageSize   types.Int64  `tfsdk:"hugepage_size"`
	HostNumaNode   types.Int64  `tfsdk:"host_numa_node"`
	HotplugSize    types.Int64  `tfsdk:"hotplug_size"`
	HotpluggedSize types.Int64  `tfsdk:"hotplugged_size"`
	Prefault       types.Bool   `tfsdk:"prefault"`
	Reserve        types.Bool   `tfsdk:"reserve"`
}

// ---------------------------------------------------------------------------
// Rate limit groups
// ---------------------------------------------------------------------------

type vmRateLimitGroupModel struct {
	ID          types.String        `tfsdk:"id"`
	RateLimiter *vmRateLimiterModel `tfsdk:"rate_limiter"`
}

type vmRateLimiterModel struct {
	Bandwidth *vmTokenBucketModel `tfsdk:"bandwidth"`
	Ops       *vmTokenBucketModel `tfsdk:"ops"`
}

type vmTokenBucketModel struct {
	Size         types.Int64 `tfsdk:"size"`
	OneTimeBurst types.Int64 `tfsdk:"one_time_burst"`
	RefillTime   types.Int64 `tfsdk:"refill_time"`
}

// ---------------------------------------------------------------------------
// Disk
// ---------------------------------------------------------------------------

type vmDiskModel struct {
	Path            types.String           `tfsdk:"path"`
	Readonly        types.Bool             `tfsdk:"readonly"`
	Direct          types.Bool             `tfsdk:"direct"`
	Iommu           types.Bool             `tfsdk:"iommu"`
	NumQueues       types.Int64            `tfsdk:"num_queues"`
	QueueSize       types.Int64            `tfsdk:"queue_size"`
	VhostUser       types.Bool             `tfsdk:"vhost_user"`
	VhostSocket     types.String           `tfsdk:"vhost_socket"`
	RateLimiter     *vmRateLimiterModel    `tfsdk:"rate_limiter"`
	PCISegment      types.Int64            `tfsdk:"pci_segment"`
	PCIDeviceID     types.Int64            `tfsdk:"pci_device_id"`
	ID              types.String           `tfsdk:"id"`
	Serial          types.String           `tfsdk:"serial"`
	RateLimitGroup  types.String           `tfsdk:"rate_limit_group"`
	QueueAffinity   []vmQueueAffinityModel `tfsdk:"queue_affinity"`
	BackingFiles    types.Bool             `tfsdk:"backing_files"`
	Sparse          types.Bool             `tfsdk:"sparse"`
	ImageType       types.String           `tfsdk:"image_type"`
	LockGranularity types.String           `tfsdk:"lock_granularity"`
}

type vmQueueAffinityModel struct {
	QueueIndex types.Int64   `tfsdk:"queue_index"`
	HostCPUs   []types.Int64 `tfsdk:"host_cpus"`
}

// ---------------------------------------------------------------------------
// Net
// ---------------------------------------------------------------------------

type vmNetModel struct {
	Tap         types.String        `tfsdk:"tap"`
	IP          types.String        `tfsdk:"ip"`
	Mask        types.String        `tfsdk:"mask"`
	MAC         types.String        `tfsdk:"mac"`
	HostMAC     types.String        `tfsdk:"host_mac"`
	MTU         types.Int64         `tfsdk:"mtu"`
	Iommu       types.Bool          `tfsdk:"iommu"`
	NumQueues   types.Int64         `tfsdk:"num_queues"`
	QueueSize   types.Int64         `tfsdk:"queue_size"`
	VhostUser   types.Bool          `tfsdk:"vhost_user"`
	VhostSocket types.String        `tfsdk:"vhost_socket"`
	VhostMode   types.String        `tfsdk:"vhost_mode"`
	ID          types.String        `tfsdk:"id"`
	PCISegment  types.Int64         `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64         `tfsdk:"pci_device_id"`
	RateLimiter *vmRateLimiterModel `tfsdk:"rate_limiter"`
	OffloadTSO  types.Bool          `tfsdk:"offload_tso"`
	OffloadUFO  types.Bool          `tfsdk:"offload_ufo"`
	OffloadCsum types.Bool          `tfsdk:"offload_csum"`
}

// ---------------------------------------------------------------------------
// Filesystem (virtio-fs)
// ---------------------------------------------------------------------------

type vmFsModel struct {
	Tag         types.String `tfsdk:"tag"`
	Socket      types.String `tfsdk:"socket"`
	NumQueues   types.Int64  `tfsdk:"num_queues"`
	QueueSize   types.Int64  `tfsdk:"queue_size"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
	ID          types.String `tfsdk:"id"`
}

// ---------------------------------------------------------------------------
// Generic vhost-user
// ---------------------------------------------------------------------------

type vmGenericVhostUserModel struct {
	Socket      types.String `tfsdk:"socket"`
	QueueSizes  types.List   `tfsdk:"queue_sizes"`
	DeviceType  types.Int64  `tfsdk:"device_type"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
}

// ---------------------------------------------------------------------------
// PMEM
// ---------------------------------------------------------------------------

type vmPmemModel struct {
	File          types.String `tfsdk:"file"`
	Size          types.Int64  `tfsdk:"size"`
	Iommu         types.Bool   `tfsdk:"iommu"`
	DiscardWrites types.Bool   `tfsdk:"discard_writes"`
	PCISegment    types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID   types.Int64  `tfsdk:"pci_device_id"`
	ID            types.String `tfsdk:"id"`
}

// ---------------------------------------------------------------------------
// Serial / Console / Debug Console
// ---------------------------------------------------------------------------

type vmSerialModel struct {
	File   types.String `tfsdk:"file"`
	Socket types.String `tfsdk:"socket"`
	Mode   types.String `tfsdk:"mode"`
}

type vmConsoleModel struct {
	File        types.String `tfsdk:"file"`
	Socket      types.String `tfsdk:"socket"`
	Mode        types.String `tfsdk:"mode"`
	Iommu       types.Bool   `tfsdk:"iommu"`
	ID          types.String `tfsdk:"id"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
}

type vmDebugConsoleModel struct {
	File   types.String `tfsdk:"file"`
	Mode   types.String `tfsdk:"mode"`
	IOBase types.Int64  `tfsdk:"iobase"`
}

// ---------------------------------------------------------------------------
// Device (VFIO)
// ---------------------------------------------------------------------------

type vmDeviceModel struct {
	Path               types.String `tfsdk:"path"`
	Iommu              types.Bool   `tfsdk:"iommu"`
	PCISegment         types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID        types.Int64  `tfsdk:"pci_device_id"`
	ID                 types.String `tfsdk:"id"`
	XNvGPUDirectClique types.Int64  `tfsdk:"x_nv_gpudirect_clique"`
}

// ---------------------------------------------------------------------------
// User Device
// ---------------------------------------------------------------------------

type vmUserDeviceModel struct {
	Socket      types.String `tfsdk:"socket"`
	ID          types.String `tfsdk:"id"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
}

// ---------------------------------------------------------------------------
// vDPA
// ---------------------------------------------------------------------------

type vmVdpaModel struct {
	Path        types.String `tfsdk:"path"`
	NumQueues   types.Int64  `tfsdk:"num_queues"`
	Iommu       types.Bool   `tfsdk:"iommu"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
	ID          types.String `tfsdk:"id"`
}

// ---------------------------------------------------------------------------
// Vsock
// ---------------------------------------------------------------------------

type vmVsockModel struct {
	CID         types.Int64  `tfsdk:"cid"`
	Socket      types.String `tfsdk:"socket"`
	Iommu       types.Bool   `tfsdk:"iommu"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
	ID          types.String `tfsdk:"id"`
}

// ---------------------------------------------------------------------------
// NUMA
// ---------------------------------------------------------------------------

type vmNumaModel struct {
	GuestNumaID types.Int64           `tfsdk:"guest_numa_id"`
	CPUs        []types.Int64         `tfsdk:"cpus"`
	Distances   []vmNumaDistanceModel `tfsdk:"distances"`
	MemoryZones []types.String        `tfsdk:"memory_zones"`
	PCISegments []types.Int64         `tfsdk:"pci_segments"`
	DeviceID    types.String          `tfsdk:"device_id"`
}

type vmNumaDistanceModel struct {
	Destination types.Int64 `tfsdk:"destination"`
	Distance    types.Int64 `tfsdk:"distance"`
}

// ---------------------------------------------------------------------------
// RNG
// ---------------------------------------------------------------------------

type vmRngModel struct {
	Src         types.String `tfsdk:"src"`
	ID          types.String `tfsdk:"id"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
	Iommu       types.Bool   `tfsdk:"iommu"`
}

// ---------------------------------------------------------------------------
// Balloon
// ---------------------------------------------------------------------------

type vmBalloonModel struct {
	Size              types.Int64  `tfsdk:"size"`
	ID                types.String `tfsdk:"id"`
	PCISegment        types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID       types.Int64  `tfsdk:"pci_device_id"`
	Iommu             types.Bool   `tfsdk:"iommu"`
	DeflateOnOOM      types.Bool   `tfsdk:"deflate_on_oom"`
	FreePageReporting types.Bool   `tfsdk:"free_page_reporting"`
}

// ---------------------------------------------------------------------------
// RTC
// ---------------------------------------------------------------------------

type vmRtcModel struct {
	ID          types.String `tfsdk:"id"`
	PCISegment  types.Int64  `tfsdk:"pci_segment"`
	PCIDeviceID types.Int64  `tfsdk:"pci_device_id"`
	Iommu       types.Bool   `tfsdk:"iommu"`
}

// ---------------------------------------------------------------------------
// PCI Segment
// ---------------------------------------------------------------------------

type vmPciSegmentModel struct {
	PCISegment           types.Int64 `tfsdk:"pci_segment"`
	MMIO32ApertureWeight types.Int64 `tfsdk:"mmio32_aperture_weight"`
	MMIO64ApertureWeight types.Int64 `tfsdk:"mmio64_aperture_weight"`
}

// ---------------------------------------------------------------------------
// Platform
// ---------------------------------------------------------------------------

type vmPlatformModel struct {
	NumPCISegments        types.Int64  `tfsdk:"num_pci_segments"`
	IOMMUSegments         types.List   `tfsdk:"iommu_segments"`
	IOMMUAddressWidthBits types.Int64  `tfsdk:"iommu_address_width_bits"`
	SystemSerialNumber    types.String `tfsdk:"system_serial_number"`
	SerialNumber          types.String `tfsdk:"serial_number"`
	SystemUUID            types.String `tfsdk:"system_uuid"`
	UUID                  types.String `tfsdk:"uuid"`
	OEMStrings            types.List   `tfsdk:"oem_strings"`
	SystemManufacturer    types.String `tfsdk:"system_manufacturer"`
	SystemProductName     types.String `tfsdk:"system_product_name"`
	SystemVersion         types.String `tfsdk:"system_version"`
	SystemFamily          types.String `tfsdk:"system_family"`
	SystemSKUNumber       types.String `tfsdk:"system_sku_number"`
	ChassisAssetTag       types.String `tfsdk:"chassis_asset_tag"`
	Tdx                   types.Bool   `tfsdk:"tdx"`
	SevSnp                types.Bool   `tfsdk:"sev_snp"`
	Iommufd               types.Bool   `tfsdk:"iommufd"`
	VfioP2PDma            types.Bool   `tfsdk:"vfio_p2p_dma"`
}

// ---------------------------------------------------------------------------
// TPM
// ---------------------------------------------------------------------------

type vmTpmModel struct {
	Socket types.String `tfsdk:"socket"`
}

// ---------------------------------------------------------------------------
// Landlock
// ---------------------------------------------------------------------------

type vmLandlockRuleModel struct {
	Path   types.String `tfsdk:"path"`
	Access types.String `tfsdk:"access"`
}

// ---------------------------------------------------------------------------
// Conversion helpers: Terraform model → client types
// ---------------------------------------------------------------------------

// toClientConfig converts the Terraform model to a client.VmConfig for API calls.
func (m *vmResourceModel) toClientConfig() *client.VmConfig {
	cfg := &client.VmConfig{}

	if m.Payload != nil {
		cfg.Payload = &client.PayloadConfig{
			Firmware:  m.Payload.Firmware.ValueString(),
			Kernel:    m.Payload.Kernel.ValueString(),
			Cmdline:   m.Payload.Cmdline.ValueString(),
			Initramfs: m.Payload.Initramfs.ValueString(),
			IGVM:      m.Payload.IGVM.ValueString(),
			HostData:  m.Payload.HostData.ValueString(),
		}
	}

	if m.Cpus != nil {
		cfg.Cpus = &client.CpusConfig{
			BootVcpus:      int(m.Cpus.BootVcpus.ValueInt64()),
			MaxVcpus:       int(m.Cpus.MaxVcpus.ValueInt64()),
			KvmHyperv:      optionalBool(m.Cpus.KvmHyperv),
			MaxPhysBits:    optionalInt(m.Cpus.MaxPhysBits),
			Nested:         optionalBool(m.Cpus.Nested),
			CoreScheduling: optionalCoreScheduling(m.Cpus.CoreScheduling),
		}
		if m.Cpus.Topology != nil {
			cfg.Cpus.Topology = &client.CpuTopology{
				ThreadsPerCore: optionalInt(m.Cpus.Topology.ThreadsPerCore),
				CoresPerDie:    optionalInt(m.Cpus.Topology.CoresPerDie),
				DiesPerPackage: optionalInt(m.Cpus.Topology.DiesPerPackage),
				Packages:       optionalInt(m.Cpus.Topology.Packages),
			}
		}
		cfg.Cpus.Affinity = convertCpuAffinities(m.Cpus.Affinity)
		if m.Cpus.Features != nil {
			cfg.Cpus.Features = &client.CpuFeatures{
				Amx: optionalBool(m.Cpus.Features.Amx),
			}
		}
	}

	if m.Memory != nil {
		cfg.Memory = &client.MemoryConfig{
			Size:           m.Memory.Size.ValueInt64(),
			HotplugSize:    optionalInt64(m.Memory.HotplugSize),
			HotpluggedSize: optionalInt64(m.Memory.HotpluggedSize),
			Mergeable:      optionalBool(m.Memory.Mergeable),
			HotplugMethod:  m.Memory.HotplugMethod.ValueString(),
			Shared:         optionalBool(m.Memory.Shared),
			Hugepages:      optionalBool(m.Memory.Hugepages),
			HugepageSize:   optionalInt64(m.Memory.HugepageSize),
			Prefault:       optionalBool(m.Memory.Prefault),
			Reserve:        optionalBool(m.Memory.Reserve),
			Thp:            optionalBool(m.Memory.Thp),
		}
		cfg.Memory.Zones = convertMemoryZones(m.Memory.Zones)
	}

	cfg.RateLimitGroups = convertRateLimitGroups(m.RateLimitGroups)
	cfg.Disks = convertDisks(m.Disks)
	cfg.Net = convertNets(m.Net)
	cfg.FS = convertFs(m.FS)
	cfg.GenericVhostUser = convertGenericVhostUser(m.GenericVhostUser)
	cfg.Pmem = convertPmem(m.Pmem)

	if m.Serial != nil {
		cfg.Serial = &client.SerialConfig{
			File:   m.Serial.File.ValueString(),
			Socket: m.Serial.Socket.ValueString(),
			Mode:   client.ConsoleMode(m.Serial.Mode.ValueString()),
		}
	}

	if m.Console != nil {
		cfg.Console = &client.ConsoleConfig{
			File:        m.Console.File.ValueString(),
			Socket:      m.Console.Socket.ValueString(),
			Mode:        client.ConsoleMode(m.Console.Mode.ValueString()),
			Iommu:       optionalBool(m.Console.Iommu),
			ID:          m.Console.ID.ValueString(),
			PCISegment:  optionalInt16(m.Console.PCISegment),
			PCIDeviceID: optionalUint8(m.Console.PCIDeviceID),
		}
	}

	if m.DebugConsole != nil {
		cfg.DebugConsole = &client.DebugConsoleConfig{
			File:   m.DebugConsole.File.ValueString(),
			Mode:   client.ConsoleMode(m.DebugConsole.Mode.ValueString()),
			IOBase: optionalInt(m.DebugConsole.IOBase),
		}
	}

	cfg.Devices = convertDevices(m.Devices)
	cfg.UserDevices = convertUserDevices(m.UserDevices)
	cfg.Vdpa = convertVdpa(m.Vdpa)

	if m.Vsock != nil {
		cfg.Vsock = &client.VsockConfig{
			CID:         m.Vsock.CID.ValueInt64(),
			Socket:      m.Vsock.Socket.ValueString(),
			Iommu:       optionalBool(m.Vsock.Iommu),
			PCISegment:  optionalInt16(m.Vsock.PCISegment),
			PCIDeviceID: optionalUint8(m.Vsock.PCIDeviceID),
			ID:          m.Vsock.ID.ValueString(),
		}
	}

	cfg.Numa = convertNuma(m.Numa)

	if m.Rng != nil {
		cfg.Rng = &client.RngConfig{
			Src:         m.Rng.Src.ValueString(),
			ID:          m.Rng.ID.ValueString(),
			PCISegment:  optionalInt16(m.Rng.PCISegment),
			PCIDeviceID: optionalUint8(m.Rng.PCIDeviceID),
			Iommu:       optionalBool(m.Rng.Iommu),
		}
	}

	if m.Balloon != nil {
		cfg.Balloon = &client.BalloonConfig{
			Size:              m.Balloon.Size.ValueInt64(),
			ID:                m.Balloon.ID.ValueString(),
			PCISegment:        optionalInt16(m.Balloon.PCISegment),
			PCIDeviceID:       optionalUint8(m.Balloon.PCIDeviceID),
			Iommu:             optionalBool(m.Balloon.Iommu),
			DeflateOnOOM:      optionalBool(m.Balloon.DeflateOnOOM),
			FreePageReporting: optionalBool(m.Balloon.FreePageReporting),
		}
	}

	if m.Rtc != nil {
		cfg.Rtc = &client.RtcConfig{
			ID:          m.Rtc.ID.ValueString(),
			PCISegment:  optionalInt16(m.Rtc.PCISegment),
			PCIDeviceID: optionalUint8(m.Rtc.PCIDeviceID),
			Iommu:       optionalBool(m.Rtc.Iommu),
		}
	}

	cfg.Pvpanic = optionalBool(m.Pvpanic)
	cfg.Watchdog = optionalBool(m.Watchdog)
	cfg.Iommu = optionalBool(m.Iommu)

	cfg.PCISegments = convertPciSegments(m.PCISegments)

	if m.Platform != nil {
		cfg.Platform = &client.PlatformConfig{
			NumPCISegments:        optionalInt16(m.Platform.NumPCISegments),
			IOMMUAddressWidthBits: optionalUint8(m.Platform.IOMMUAddressWidthBits),
			SystemSerialNumber:    m.Platform.SystemSerialNumber.ValueString(),
			SerialNumber:          m.Platform.SerialNumber.ValueString(),
			SystemUUID:            m.Platform.SystemUUID.ValueString(),
			UUID:                  m.Platform.UUID.ValueString(),
			SystemManufacturer:    m.Platform.SystemManufacturer.ValueString(),
			SystemProductName:     m.Platform.SystemProductName.ValueString(),
			SystemVersion:         m.Platform.SystemVersion.ValueString(),
			SystemFamily:          m.Platform.SystemFamily.ValueString(),
			SystemSKUNumber:       m.Platform.SystemSKUNumber.ValueString(),
			ChassisAssetTag:       m.Platform.ChassisAssetTag.ValueString(),
			Tdx:                   optionalBool(m.Platform.Tdx),
			SevSnp:                optionalBool(m.Platform.SevSnp),
			Iommufd:               optionalBool(m.Platform.Iommufd),
			VfioP2PDma:            optionalBool(m.Platform.VfioP2PDma),
		}
	}

	if m.Tpm != nil {
		cfg.Tpm = &client.TpmConfig{
			Socket: m.Tpm.Socket.ValueString(),
		}
	}

	cfg.LandlockEnable = optionalBool(m.LandlockEnable)
	cfg.LandlockRules = convertLandlockRules(m.LandlockRules)

	return cfg
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func optionalBool(v types.Bool) *bool {
	if v.IsNull() {
		return nil
	}
	val := v.ValueBool()
	return &val
}

func optionalInt(v types.Int64) *int {
	if v.IsNull() {
		return nil
	}
	val := int(v.ValueInt64())
	return &val
}

func optionalInt64(v types.Int64) *int64 {
	if v.IsNull() {
		return nil
	}
	val := v.ValueInt64()
	return &val
}

func optionalInt16(v types.Int64) *int16 {
	if v.IsNull() {
		return nil
	}
	val := int16(v.ValueInt64())
	return &val
}

func optionalUint8(v types.Int64) *uint8 {
	if v.IsNull() {
		return nil
	}
	val := uint8(v.ValueInt64())
	return &val
}

func optionalCoreScheduling(v types.String) *client.CoreSchedulingMode {
	if v.IsNull() {
		return nil
	}
	val := client.CoreSchedulingMode(v.ValueString())
	return &val
}

// ---------------------------------------------------------------------------
// Slice conversion helpers
// ---------------------------------------------------------------------------

func convertCpuAffinities(in []vmCpuAffinityModel) []client.CpuAffinity {
	if in == nil {
		return nil
	}
	out := make([]client.CpuAffinity, len(in))
	for i, v := range in {
		var hostCPUs []int
		if len(v.HostCPUs) > 0 {
			hostCPUs = make([]int, len(v.HostCPUs))
			for j, h := range v.HostCPUs {
				hostCPUs[j] = int(h.ValueInt64())
			}
		}
		out[i] = client.CpuAffinity{
			Vcpu:     int(v.Vcpu.ValueInt64()),
			HostCPUs: hostCPUs,
		}
	}
	return out
}

func convertMemoryZones(in []vmMemoryZoneModel) []client.MemoryZoneConfig {
	if in == nil {
		return nil
	}
	out := make([]client.MemoryZoneConfig, len(in))
	for i, v := range in {
		out[i] = client.MemoryZoneConfig{
			ID:             v.ID.ValueString(),
			Size:           v.Size.ValueInt64(),
			File:           v.File.ValueString(),
			Mergeable:      optionalBool(v.Mergeable),
			Shared:         optionalBool(v.Shared),
			Hugepages:      optionalBool(v.Hugepages),
			HugepageSize:   optionalInt64(v.HugepageSize),
			HostNumaNode:   optionalInt32(v.HostNumaNode),
			HotplugSize:    optionalInt64(v.HotplugSize),
			HotpluggedSize: optionalInt64(v.HotpluggedSize),
			Prefault:       optionalBool(v.Prefault),
			Reserve:        optionalBool(v.Reserve),
		}
	}
	return out
}

func optionalInt32(v types.Int64) *int32 {
	if v.IsNull() {
		return nil
	}
	val := int32(v.ValueInt64())
	return &val
}

func convertRateLimitGroups(in []vmRateLimitGroupModel) []client.RateLimitGroupConfig {
	if in == nil {
		return nil
	}
	out := make([]client.RateLimitGroupConfig, len(in))
	for i, v := range in {
		out[i] = client.RateLimitGroupConfig{
			ID:                v.ID.ValueString(),
			RateLimiterConfig: convertRateLimiter(v.RateLimiter),
		}
	}
	return out
}

func convertRateLimiter(in *vmRateLimiterModel) *client.RateLimiterConfig {
	if in == nil {
		return nil
	}
	return &client.RateLimiterConfig{
		Bandwidth: convertTokenBucket(in.Bandwidth),
		Ops:       convertTokenBucket(in.Ops),
	}
}

func convertTokenBucket(in *vmTokenBucketModel) *client.TokenBucket {
	if in == nil {
		return nil
	}
	return &client.TokenBucket{
		Size:         in.Size.ValueInt64(),
		OneTimeBurst: in.OneTimeBurst.ValueInt64(),
		RefillTime:   in.RefillTime.ValueInt64(),
	}
}

func convertDisks(in []vmDiskModel) []client.DiskConfig {
	if in == nil {
		return nil
	}
	out := make([]client.DiskConfig, len(in))
	for i, v := range in {
		out[i] = client.DiskConfig{
			Path:              v.Path.ValueString(),
			Readonly:          optionalBool(v.Readonly),
			Direct:            optionalBool(v.Direct),
			Iommu:             optionalBool(v.Iommu),
			NumQueues:         optionalInt(v.NumQueues),
			QueueSize:         optionalInt(v.QueueSize),
			VhostUser:         optionalBool(v.VhostUser),
			VhostSocket:       v.VhostSocket.ValueString(),
			RateLimiterConfig: convertRateLimiter(v.RateLimiter),
			PCISegment:        optionalInt16(v.PCISegment),
			PCIDeviceID:       optionalUint8(v.PCIDeviceID),
			ID:                v.ID.ValueString(),
			Serial:            v.Serial.ValueString(),
			RateLimitGroup:    v.RateLimitGroup.ValueString(),
			QueueAffinity:     convertQueueAffinities(v.QueueAffinity),
			BackingFiles:      optionalBool(v.BackingFiles),
			Sparse:            optionalBool(v.Sparse),
			ImageType:         optionalImageType(v.ImageType),
			LockGranularity:   optionalLockGranularity(v.LockGranularity),
		}
	}
	return out
}

func optionalImageType(v types.String) *client.ImageType {
	if v.IsNull() {
		return nil
	}
	val := client.ImageType(v.ValueString())
	return &val
}

func optionalLockGranularity(v types.String) *client.LockGranularity {
	if v.IsNull() {
		return nil
	}
	val := client.LockGranularity(v.ValueString())
	return &val
}

func convertQueueAffinities(in []vmQueueAffinityModel) []client.VirtQueueAffinity {
	if in == nil {
		return nil
	}
	out := make([]client.VirtQueueAffinity, len(in))
	for i, v := range in {
		var hostCPUs []int
		if len(v.HostCPUs) > 0 {
			hostCPUs = make([]int, len(v.HostCPUs))
			for j, h := range v.HostCPUs {
				hostCPUs[j] = int(h.ValueInt64())
			}
		}
		out[i] = client.VirtQueueAffinity{
			QueueIndex: int(v.QueueIndex.ValueInt64()),
			HostCPUs:   hostCPUs,
		}
	}
	return out
}

func convertNets(in []vmNetModel) []client.NetConfig {
	if in == nil {
		return nil
	}
	out := make([]client.NetConfig, len(in))
	for i, v := range in {
		out[i] = client.NetConfig{
			Tap:               v.Tap.ValueString(),
			IP:                v.IP.ValueString(),
			Mask:              v.Mask.ValueString(),
			MAC:               v.MAC.ValueString(),
			HostMAC:           v.HostMAC.ValueString(),
			MTU:               optionalInt(v.MTU),
			Iommu:             optionalBool(v.Iommu),
			NumQueues:         optionalInt(v.NumQueues),
			QueueSize:         optionalInt(v.QueueSize),
			VhostUser:         optionalBool(v.VhostUser),
			VhostSocket:       v.VhostSocket.ValueString(),
			VhostMode:         v.VhostMode.ValueString(),
			ID:                v.ID.ValueString(),
			PCISegment:        optionalInt16(v.PCISegment),
			PCIDeviceID:       optionalUint8(v.PCIDeviceID),
			RateLimiterConfig: convertRateLimiter(v.RateLimiter),
			OffloadTSO:        optionalBool(v.OffloadTSO),
			OffloadUFO:        optionalBool(v.OffloadUFO),
			OffloadCsum:       optionalBool(v.OffloadCsum),
		}
	}
	return out
}

func convertFs(in []vmFsModel) []client.FsConfig {
	if in == nil {
		return nil
	}
	out := make([]client.FsConfig, len(in))
	for i, v := range in {
		out[i] = client.FsConfig{
			Tag:         v.Tag.ValueString(),
			Socket:      v.Socket.ValueString(),
			NumQueues:   int(v.NumQueues.ValueInt64()),
			QueueSize:   int(v.QueueSize.ValueInt64()),
			PCISegment:  optionalInt16(v.PCISegment),
			PCIDeviceID: optionalUint8(v.PCIDeviceID),
			ID:          v.ID.ValueString(),
		}
	}
	return out
}

func convertGenericVhostUser(in []vmGenericVhostUserModel) []client.GenericVhostUserConfig {
	if in == nil {
		return nil
	}
	out := make([]client.GenericVhostUserConfig, len(in))
	for i, v := range in {
		out[i] = client.GenericVhostUserConfig{
			Socket:      v.Socket.ValueString(),
			DeviceType:  uint32(v.DeviceType.ValueInt64()),
			PCISegment:  optionalInt16(v.PCISegment),
			PCIDeviceID: optionalUint8(v.PCIDeviceID),
		}
	}
	return out
}

func convertPmem(in []vmPmemModel) []client.PmemConfig {
	if in == nil {
		return nil
	}
	out := make([]client.PmemConfig, len(in))
	for i, v := range in {
		out[i] = client.PmemConfig{
			File:          v.File.ValueString(),
			Size:          optionalInt64(v.Size),
			Iommu:         optionalBool(v.Iommu),
			DiscardWrites: optionalBool(v.DiscardWrites),
			PCISegment:    optionalInt16(v.PCISegment),
			PCIDeviceID:   optionalUint8(v.PCIDeviceID),
			ID:            v.ID.ValueString(),
		}
	}
	return out
}

func convertDevices(in []vmDeviceModel) []client.DeviceConfig {
	if in == nil {
		return nil
	}
	out := make([]client.DeviceConfig, len(in))
	for i, v := range in {
		out[i] = client.DeviceConfig{
			Path:               v.Path.ValueString(),
			Iommu:              optionalBool(v.Iommu),
			PCISegment:         optionalInt16(v.PCISegment),
			PCIDeviceID:        optionalUint8(v.PCIDeviceID),
			ID:                 v.ID.ValueString(),
			XNvGPUDirectClique: optionalInt8(v.XNvGPUDirectClique),
		}
	}
	return out
}

func optionalInt8(v types.Int64) *int8 {
	if v.IsNull() {
		return nil
	}
	val := int8(v.ValueInt64())
	return &val
}

func convertUserDevices(in []vmUserDeviceModel) []client.UserDeviceConfig {
	if in == nil {
		return nil
	}
	out := make([]client.UserDeviceConfig, len(in))
	for i, v := range in {
		out[i] = client.UserDeviceConfig{
			Socket:      v.Socket.ValueString(),
			ID:          v.ID.ValueString(),
			PCISegment:  optionalInt16(v.PCISegment),
			PCIDeviceID: optionalUint8(v.PCIDeviceID),
		}
	}
	return out
}

func convertVdpa(in []vmVdpaModel) []client.VdpaConfig {
	if in == nil {
		return nil
	}
	out := make([]client.VdpaConfig, len(in))
	for i, v := range in {
		out[i] = client.VdpaConfig{
			Path:        v.Path.ValueString(),
			NumQueues:   int(v.NumQueues.ValueInt64()),
			Iommu:       optionalBool(v.Iommu),
			PCISegment:  optionalInt16(v.PCISegment),
			PCIDeviceID: optionalUint8(v.PCIDeviceID),
			ID:          v.ID.ValueString(),
		}
	}
	return out
}

func convertNuma(in []vmNumaModel) []client.NumaConfig {
	if in == nil {
		return nil
	}
	out := make([]client.NumaConfig, len(in))
	for i, v := range in {
		var cpus []int32
		if len(v.CPUs) > 0 {
			cpus = make([]int32, len(v.CPUs))
			for j, c := range v.CPUs {
				cpus[j] = int32(c.ValueInt64())
			}
		}
		var memoryZones []string
		if len(v.MemoryZones) > 0 {
			memoryZones = make([]string, len(v.MemoryZones))
			for j, z := range v.MemoryZones {
				memoryZones[j] = z.ValueString()
			}
		}
		var pciSegments []int32
		if len(v.PCISegments) > 0 {
			pciSegments = make([]int32, len(v.PCISegments))
			for j, s := range v.PCISegments {
				pciSegments[j] = int32(s.ValueInt64())
			}
		}
		var distances []client.NumaDistance
		if len(v.Distances) > 0 {
			distances = make([]client.NumaDistance, len(v.Distances))
			for j, d := range v.Distances {
				distances[j] = client.NumaDistance{
					Destination: int32(d.Destination.ValueInt64()),
					Distance:    int32(d.Distance.ValueInt64()),
				}
			}
		}
		out[i] = client.NumaConfig{
			GuestNumaID: int32(v.GuestNumaID.ValueInt64()),
			CPUs:        cpus,
			Distances:   distances,
			MemoryZones: memoryZones,
			PCISegments: pciSegments,
			DeviceID:    v.DeviceID.ValueString(),
		}
	}
	return out
}

func convertPciSegments(in []vmPciSegmentModel) []client.PciSegmentConfig {
	if in == nil {
		return nil
	}
	out := make([]client.PciSegmentConfig, len(in))
	for i, v := range in {
		out[i] = client.PciSegmentConfig{
			PCISegment:           int16(v.PCISegment.ValueInt64()),
			MMIO32ApertureWeight: optionalInt32(v.MMIO32ApertureWeight),
			MMIO64ApertureWeight: optionalInt32(v.MMIO64ApertureWeight),
		}
	}
	return out
}

func convertLandlockRules(in []vmLandlockRuleModel) []client.LandlockConfig {
	if in == nil {
		return nil
	}
	out := make([]client.LandlockConfig, len(in))
	for i, v := range in {
		out[i] = client.LandlockConfig{
			Path:   v.Path.ValueString(),
			Access: v.Access.ValueString(),
		}
	}
	return out
}

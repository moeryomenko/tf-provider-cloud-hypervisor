package client

// ---------------------------------------------------------------------------
// Top-level response types
// ---------------------------------------------------------------------------

// VmmPingResponse is the response from GET /vmm.ping.
type VmmPingResponse struct {
	BuildVersion string   `json:"build_version,omitempty"`
	Version      string   `json:"version"`
	PID          int64    `json:"pid,omitempty"`
	Features     []string `json:"features,omitempty"`
}

// VmState represents the lifecycle state of a VM instance.
type VmState string

const (
	VmStateCreated  VmState = "Created"
	VmStateRunning  VmState = "Running"
	VmStateShutdown VmState = "Shutdown"
	VmStatePaused   VmState = "Paused"
)

// VmInfo is the response from GET /vm.info.
type VmInfo struct {
	Config           *VmConfig             `json:"config"`
	State            VmState               `json:"state"`
	MemoryActualSize int64                 `json:"memory_actual_size,omitempty"`
	DeviceTree       map[string]DeviceNode `json:"device_tree,omitempty"`
}

// DeviceNode represents a device in the device tree.
type DeviceNode struct {
	ID        string           `json:"id,omitempty"`
	Resources []map[string]any `json:"resources,omitempty"`
	Children  []string         `json:"children,omitempty"`
	PciBDF    string           `json:"pci_bdf,omitempty"`
}

// PciDeviceInfo is the response from hotplug endpoints (200 case).
type PciDeviceInfo struct {
	ID  string `json:"id"`
	BDF string `json:"bdf"`
}

// VmRemoveDevice is the request body for PUT /vm.remove-device.
type VmRemoveDevice struct {
	ID string `json:"id"`
}

// ---------------------------------------------------------------------------
// VmConfig and nested configuration types
// ---------------------------------------------------------------------------

// VmConfig is the root VM configuration (only payload is required).
type VmConfig struct {
	Payload          *PayloadConfig           `json:"payload"`
	Cpus             *CpusConfig              `json:"cpus,omitempty"`
	Memory           *MemoryConfig            `json:"memory,omitempty"`
	RateLimitGroups  []RateLimitGroupConfig   `json:"rate_limit_groups,omitempty"`
	Disks            []DiskConfig             `json:"disks,omitempty"`
	Net              []NetConfig              `json:"net,omitempty"`
	Rng              *RngConfig               `json:"rng,omitempty"`
	Balloon          *BalloonConfig           `json:"balloon,omitempty"`
	FS               []FsConfig               `json:"fs,omitempty"`
	GenericVhostUser []GenericVhostUserConfig `json:"generic-vhost-user,omitempty"`
	Pmem             []PmemConfig             `json:"pmem,omitempty"`
	Serial           *SerialConfig            `json:"serial,omitempty"`
	Console          *ConsoleConfig           `json:"console,omitempty"`
	DebugConsole     *DebugConsoleConfig      `json:"debug_console,omitempty"`
	Devices          []DeviceConfig           `json:"devices,omitempty"`
	UserDevices      []UserDeviceConfig       `json:"user_devices,omitempty"`
	Vdpa             []VdpaConfig             `json:"vdpa,omitempty"`
	Vsock            *VsockConfig             `json:"vsock,omitempty"`
	Numa             []NumaConfig             `json:"numa,omitempty"`
	Iommu            *bool                    `json:"iommu,omitempty"`
	Watchdog         *bool                    `json:"watchdog,omitempty"`
	Rtc              *RtcConfig               `json:"rtc,omitempty"`
	Pvpanic          *bool                    `json:"pvpanic,omitempty"`
	PCISegments      []PciSegmentConfig       `json:"pci_segments,omitempty"`
	Platform         *PlatformConfig          `json:"platform,omitempty"`
	Tpm              *TpmConfig               `json:"tpm,omitempty"`
	LandlockEnable   *bool                    `json:"landlock_enable,omitempty"`
	LandlockRules    []LandlockConfig         `json:"landlock_rules,omitempty"`
}

// PayloadConfig describes the payload (kernel/initrd/cmdline) to boot.
type PayloadConfig struct {
	Firmware  string `json:"firmware,omitempty"`
	Kernel    string `json:"kernel,omitempty"`
	Cmdline   string `json:"cmdline,omitempty"`
	Initramfs string `json:"initramfs,omitempty"`
	IGVM      string `json:"igvm,omitempty"`
	HostData  string `json:"host_data,omitempty"`
}

// ---------------------------------------------------------------------------
// CPU configuration
// ---------------------------------------------------------------------------

// CpusConfig defines the CPU topology and features for the VM.
type CpusConfig struct {
	BootVcpus      int                 `json:"boot_vcpus"`
	MaxVcpus       int                 `json:"max_vcpus"`
	Topology       *CpuTopology        `json:"topology,omitempty"`
	KvmHyperv      *bool               `json:"kvm_hyperv,omitempty"`
	MaxPhysBits    *int                `json:"max_phys_bits,omitempty"`
	Nested         *bool               `json:"nested,omitempty"`
	Affinity       []CpuAffinity       `json:"affinity,omitempty"`
	Features       *CpuFeatures        `json:"features,omitempty"`
	CoreScheduling *CoreSchedulingMode `json:"core_scheduling,omitempty"`
}

// CpuTopology describes the CPU topology.
type CpuTopology struct {
	ThreadsPerCore *int `json:"threads_per_core,omitempty"`
	CoresPerDie    *int `json:"cores_per_die,omitempty"`
	DiesPerPackage *int `json:"dies_per_package,omitempty"`
	Packages       *int `json:"packages,omitempty"`
}

// CpuAffinity binds a vCPU to a set of host CPUs.
type CpuAffinity struct {
	Vcpu     int   `json:"vcpu"`
	HostCPUs []int `json:"host_cpus"`
}

// CpuFeatures enables or disables CPU feature flags.
type CpuFeatures struct {
	Amx *bool `json:"amx,omitempty"`
}

// CoreSchedulingMode controls core scheduling for the VM.
type CoreSchedulingMode string

// ---------------------------------------------------------------------------
// Memory configuration
// ---------------------------------------------------------------------------

// MemoryConfig defines the memory layout for the VM.
type MemoryConfig struct {
	Size           int64              `json:"size"`
	HotplugSize    *int64             `json:"hotplug_size,omitempty"`
	HotpluggedSize *int64             `json:"hotplugged_size,omitempty"`
	Mergeable      *bool              `json:"mergeable,omitempty"`
	HotplugMethod  string             `json:"hotplug_method,omitempty"`
	Shared         *bool              `json:"shared,omitempty"`
	Hugepages      *bool              `json:"hugepages,omitempty"`
	HugepageSize   *int64             `json:"hugepage_size,omitempty"`
	Prefault       *bool              `json:"prefault,omitempty"`
	Reserve        *bool              `json:"reserve,omitempty"`
	Thp            *bool              `json:"thp,omitempty"`
	Zones          []MemoryZoneConfig `json:"zones,omitempty"`
}

// MemoryZoneConfig defines a memory zone in the VM.
type MemoryZoneConfig struct {
	ID             string `json:"id"`
	Size           int64  `json:"size"`
	File           string `json:"file,omitempty"`
	Mergeable      *bool  `json:"mergeable,omitempty"`
	Shared         *bool  `json:"shared,omitempty"`
	Hugepages      *bool  `json:"hugepages,omitempty"`
	HugepageSize   *int64 `json:"hugepage_size,omitempty"`
	HostNumaNode   *int32 `json:"host_numa_node,omitempty"`
	HotplugSize    *int64 `json:"hotplug_size,omitempty"`
	HotpluggedSize *int64 `json:"hotplugged_size,omitempty"`
	Prefault       *bool  `json:"prefault,omitempty"`
	Reserve        *bool  `json:"reserve,omitempty"`
}

// ---------------------------------------------------------------------------
// Rate limiting
// ---------------------------------------------------------------------------

// TokenBucket defines a token bucket for rate limiting.
type TokenBucket struct {
	Size         int64 `json:"size"`
	OneTimeBurst int64 `json:"one_time_burst,omitempty"`
	RefillTime   int64 `json:"refill_time"`
}

// RateLimiterConfig defines IO rate limits with independent bandwidth and ops
// token buckets.
type RateLimiterConfig struct {
	Bandwidth *TokenBucket `json:"bandwidth,omitempty"`
	Ops       *TokenBucket `json:"ops,omitempty"`
}

// RateLimitGroupConfig defines a rate limit group that can be referenced by
// other devices.
type RateLimitGroupConfig struct {
	ID                string             `json:"id"`
	RateLimiterConfig *RateLimiterConfig `json:"rate_limiter_config"`
}

// ---------------------------------------------------------------------------
// Disk configuration
// ---------------------------------------------------------------------------

// VirtQueueAffinity binds a virtqueue to a set of host CPUs.
type VirtQueueAffinity struct {
	QueueIndex int   `json:"queue_index"`
	HostCPUs   []int `json:"host_cpus"`
}

// ImageType describes the disk image format.
type ImageType string

// LockGranularity controls the locking granularity for disk access.
type LockGranularity string

// DiskConfig defines a virtual disk device.
type DiskConfig struct {
	Path              string              `json:"path,omitempty"`
	Readonly          *bool               `json:"readonly,omitempty"`
	Direct            *bool               `json:"direct,omitempty"`
	Iommu             *bool               `json:"iommu,omitempty"`
	NumQueues         *int                `json:"num_queues,omitempty"`
	QueueSize         *int                `json:"queue_size,omitempty"`
	VhostUser         *bool               `json:"vhost_user,omitempty"`
	VhostSocket       string              `json:"vhost_socket,omitempty"`
	RateLimiterConfig *RateLimiterConfig  `json:"rate_limiter_config,omitempty"`
	PCISegment        *int16              `json:"pci_segment,omitempty"`
	PCIDeviceID       *uint8              `json:"pci_device_id,omitempty"`
	ID                string              `json:"id,omitempty"`
	Serial            string              `json:"serial,omitempty"`
	RateLimitGroup    string              `json:"rate_limit_group,omitempty"`
	QueueAffinity     []VirtQueueAffinity `json:"queue_affinity,omitempty"`
	BackingFiles      *bool               `json:"backing_files,omitempty"`
	Sparse            *bool               `json:"sparse,omitempty"`
	ImageType         *ImageType          `json:"image_type,omitempty"`
	LockGranularity   *LockGranularity    `json:"lock_granularity,omitempty"`
}

// ---------------------------------------------------------------------------
// Network configuration
// ---------------------------------------------------------------------------

// NetConfig defines a virtual network device.
type NetConfig struct {
	Tap               string             `json:"tap,omitempty"`
	IP                string             `json:"ip,omitempty"`
	Mask              string             `json:"mask,omitempty"`
	MAC               string             `json:"mac,omitempty"`
	HostMAC           string             `json:"host_mac,omitempty"`
	MTU               *int               `json:"mtu,omitempty"`
	Iommu             *bool              `json:"iommu,omitempty"`
	NumQueues         *int               `json:"num_queues,omitempty"`
	QueueSize         *int               `json:"queue_size,omitempty"`
	VhostUser         *bool              `json:"vhost_user,omitempty"`
	VhostSocket       string             `json:"vhost_socket,omitempty"`
	VhostMode         string             `json:"vhost_mode,omitempty"`
	ID                string             `json:"id,omitempty"`
	PCISegment        *int16             `json:"pci_segment,omitempty"`
	PCIDeviceID       *uint8             `json:"pci_device_id,omitempty"`
	RateLimiterConfig *RateLimiterConfig `json:"rate_limiter_config,omitempty"`
	OffloadTSO        *bool              `json:"offload_tso,omitempty"`
	OffloadUFO        *bool              `json:"offload_ufo,omitempty"`
	OffloadCsum       *bool              `json:"offload_csum,omitempty"`
}

// ---------------------------------------------------------------------------
// Filesystem (virtio-fs) configuration
// ---------------------------------------------------------------------------

// FsConfig defines a virtio-fs device.
type FsConfig struct {
	Tag         string `json:"tag"`
	Socket      string `json:"socket"`
	NumQueues   int    `json:"num_queues"`
	QueueSize   int    `json:"queue_size"`
	PCISegment  *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8 `json:"pci_device_id,omitempty"`
	ID          string `json:"id,omitempty"`
}

// ---------------------------------------------------------------------------
// Generic vhost-user configuration
// ---------------------------------------------------------------------------

// GenericVhostUserConfig defines a generic vhost-user device.
type GenericVhostUserConfig struct {
	Socket      string   `json:"socket"`
	QueueSizes  []uint16 `json:"queue_sizes"`
	DeviceType  uint32   `json:"device_type"`
	PCISegment  *int16   `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8   `json:"pci_device_id,omitempty"`
}

// ---------------------------------------------------------------------------
// PMEM configuration
// ---------------------------------------------------------------------------

// PmemConfig defines a persistent memory device.
type PmemConfig struct {
	File          string `json:"file"`
	Size          *int64 `json:"size,omitempty"`
	Iommu         *bool  `json:"iommu,omitempty"`
	DiscardWrites *bool  `json:"discard_writes,omitempty"`
	PCISegment    *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID   *uint8 `json:"pci_device_id,omitempty"`
	ID            string `json:"id,omitempty"`
}

// ---------------------------------------------------------------------------
// Serial / Console / Debug Console
// ---------------------------------------------------------------------------

// ConsoleMode defines the operating mode of a serial or console device.
type ConsoleMode string

// SerialConfig defines a serial port configuration.
type SerialConfig struct {
	File   string      `json:"file,omitempty"`
	Socket string      `json:"socket,omitempty"`
	Mode   ConsoleMode `json:"mode"`
}

// ConsoleConfig defines a console device.
type ConsoleConfig struct {
	File        string      `json:"file,omitempty"`
	Socket      string      `json:"socket,omitempty"`
	Mode        ConsoleMode `json:"mode"`
	Iommu       *bool       `json:"iommu,omitempty"`
	ID          string      `json:"id,omitempty"`
	PCISegment  *int16      `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8      `json:"pci_device_id,omitempty"`
}

// DebugConsoleConfig defines a debug console device.
type DebugConsoleConfig struct {
	File   string      `json:"file,omitempty"`
	Mode   ConsoleMode `json:"mode"`
	IOBase *int        `json:"iobase,omitempty"`
}

// ---------------------------------------------------------------------------
// Device (VFIO) configuration
// ---------------------------------------------------------------------------

// DeviceConfig defines a VFIO device passthrough.
type DeviceConfig struct {
	Path               string  `json:"path,omitempty"`
	Iommu              *bool   `json:"iommu,omitempty"`
	PCISegment         *int16  `json:"pci_segment,omitempty"`
	PCIDeviceID        *uint8  `json:"pci_device_id,omitempty"`
	ID                 string  `json:"id,omitempty"`
	XNvGPUDirectClique *int8   `json:"x_nv_gpudirect_clique,omitempty"`
	XExcludeMmapBars   []int64 `json:"x_exclude_mmap_bars,omitempty"`
}

// ---------------------------------------------------------------------------
// User device configuration
// ---------------------------------------------------------------------------

// UserDeviceConfig defines a userspace (vhost-user) device.
type UserDeviceConfig struct {
	Socket      string `json:"socket"`
	ID          string `json:"id,omitempty"`
	PCISegment  *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8 `json:"pci_device_id,omitempty"`
}

// ---------------------------------------------------------------------------
// vDPA configuration
// ---------------------------------------------------------------------------

// VdpaConfig defines a vDPA device.
type VdpaConfig struct {
	Path        string `json:"path"`
	NumQueues   int    `json:"num_queues"`
	Iommu       *bool  `json:"iommu,omitempty"`
	PCISegment  *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8 `json:"pci_device_id,omitempty"`
	ID          string `json:"id,omitempty"`
}

// ---------------------------------------------------------------------------
// vhost-user socket (vsock) configuration
// ---------------------------------------------------------------------------

// VsockConfig defines a vhost-user socket (vsock) device.
type VsockConfig struct {
	CID         int64  `json:"cid"`
	Socket      string `json:"socket"`
	Iommu       *bool  `json:"iommu,omitempty"`
	PCISegment  *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8 `json:"pci_device_id,omitempty"`
	ID          string `json:"id,omitempty"`
}

// ---------------------------------------------------------------------------
// Misc device types
// ---------------------------------------------------------------------------

// RngConfig defines a virtio-rng device.
type RngConfig struct {
	Src         string `json:"src"`
	ID          string `json:"id,omitempty"`
	PCISegment  *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8 `json:"pci_device_id,omitempty"`
	Iommu       *bool  `json:"iommu,omitempty"`
}

// BalloonConfig defines a virtio-balloon device.
type BalloonConfig struct {
	Size              int64  `json:"size"`
	ID                string `json:"id,omitempty"`
	PCISegment        *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID       *uint8 `json:"pci_device_id,omitempty"`
	Iommu             *bool  `json:"iommu,omitempty"`
	DeflateOnOOM      *bool  `json:"deflate_on_oom,omitempty"`
	FreePageReporting *bool  `json:"free_page_reporting,omitempty"`
}

// RtcConfig defines the RTC device for the VM.
type RtcConfig struct {
	ID          string `json:"id,omitempty"`
	PCISegment  *int16 `json:"pci_segment,omitempty"`
	PCIDeviceID *uint8 `json:"pci_device_id,omitempty"`
	Iommu       *bool  `json:"iommu,omitempty"`
}

// TpmConfig defines a TPM device.
type TpmConfig struct {
	Socket string `json:"socket"`
}

// ---------------------------------------------------------------------------
// NUMA configuration
// ---------------------------------------------------------------------------

// NumaDistance defines the distance between two NUMA nodes.
type NumaDistance struct {
	Destination int32 `json:"destination"`
	Distance    int32 `json:"distance"`
}

// NumaConfig defines a NUMA node in the VM.
type NumaConfig struct {
	GuestNumaID int32          `json:"guest_numa_id"`
	CPUs        []int32        `json:"cpus,omitempty"`
	Distances   []NumaDistance `json:"distances,omitempty"`
	MemoryZones []string       `json:"memory_zones,omitempty"`
	PCISegments []int32        `json:"pci_segments,omitempty"`
	DeviceID    string         `json:"device_id,omitempty"`
}

// ---------------------------------------------------------------------------
// PCI / Platform configuration
// ---------------------------------------------------------------------------

// PciSegmentConfig defines a PCI segment group.
type PciSegmentConfig struct {
	PCISegment           int16  `json:"pci_segment"`
	MMIO32ApertureWeight *int32 `json:"mmio32_aperture_weight,omitempty"`
	MMIO64ApertureWeight *int32 `json:"mmio64_aperture_weight,omitempty"`
}

// PlatformConfig defines platform-level settings.
type PlatformConfig struct {
	NumPCISegments        *int16   `json:"num_pci_segments,omitempty"`
	IOMMUSegments         []int16  `json:"iommu_segments,omitempty"`
	IOMMUAddressWidthBits *uint8   `json:"iommu_address_width_bits,omitempty"`
	SystemSerialNumber    string   `json:"system_serial_number,omitempty"`
	SerialNumber          string   `json:"serial_number,omitempty"` // deprecated
	SystemUUID            string   `json:"system_uuid,omitempty"`
	UUID                  string   `json:"uuid,omitempty"` // deprecated
	OEMStrings            []string `json:"oem_strings,omitempty"`
	SystemManufacturer    string   `json:"system_manufacturer,omitempty"`
	SystemProductName     string   `json:"system_product_name,omitempty"`
	SystemVersion         string   `json:"system_version,omitempty"`
	SystemFamily          string   `json:"system_family,omitempty"`
	SystemSKUNumber       string   `json:"system_sku_number,omitempty"`
	ChassisAssetTag       string   `json:"chassis_asset_tag,omitempty"`
	Tdx                   *bool    `json:"tdx,omitempty"`
	SevSnp                *bool    `json:"sev_snp,omitempty"`
	Iommufd               *bool    `json:"iommufd,omitempty"`
	VfioP2PDma            *bool    `json:"vfio_p2p_dma,omitempty"`
}

// ---------------------------------------------------------------------------
// Landlock configuration
// ---------------------------------------------------------------------------

// LandlockConfig defines a Landlock access rule.
type LandlockConfig struct {
	Path   string `json:"path"`
	Access string `json:"access"`
}

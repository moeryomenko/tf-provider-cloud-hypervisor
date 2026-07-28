package provider

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/moeryomenko/tf-provider-cloud-hypervisor/internal/chproc"
)

// ---------------------------------------------------------------------------
// Schema Validation Tests
// ---------------------------------------------------------------------------

// TestVMResourceSchema verifies all top-level attributes exist with
// correct types and Required/Optional/Computed flags.
func TestVMResourceSchema(t *testing.T) {
	r := newVMResource()
	var resp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes

	t.Run("payload_required", func(t *testing.T) {
		attr, ok := attrs["payload"]
		if !ok {
			t.Fatal("missing payload attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("payload must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("socket_path_computed", func(t *testing.T) {
		attr, ok := attrs["socket_path"]
		if !ok {
			t.Fatal("missing socket_path attribute")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("socket_path must be StringAttribute, got %T", attr)
		}
		if !strAttr.Computed {
			t.Error("socket_path must be Computed")
		}
	})

	t.Run("socket_dir_computed", func(t *testing.T) {
		attr, ok := attrs["socket_dir"]
		if !ok {
			t.Fatal("missing socket_dir attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("socket_dir must be StringAttribute, got %T", attr)
		}
	})

	t.Run("cpus_optional", func(t *testing.T) {
		attr, ok := attrs["cpus"]
		if !ok {
			t.Fatal("missing cpus attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("cpus must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("memory_optional", func(t *testing.T) {
		attr, ok := attrs["memory"]
		if !ok {
			t.Fatal("missing memory attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("memory must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("disks_optional_list", func(t *testing.T) {
		attr, ok := attrs["disks"]
		if !ok {
			t.Fatal("missing disks attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("disks must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("net_optional_list", func(t *testing.T) {
		attr, ok := attrs["net"]
		if !ok {
			t.Fatal("missing net attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("net must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("iommu_optional", func(t *testing.T) {
		attr, ok := attrs["iommu"]
		if !ok {
			t.Fatal("missing iommu attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("iommu must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("watchdog_optional", func(t *testing.T) {
		attr, ok := attrs["watchdog"]
		if !ok {
			t.Fatal("missing watchdog attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("watchdog must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("serial_optional", func(t *testing.T) {
		attr, ok := attrs["serial"]
		if !ok {
			t.Fatal("missing serial attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("serial must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("console_optional", func(t *testing.T) {
		attr, ok := attrs["console"]
		if !ok {
			t.Fatal("missing console attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("console must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("vsock_optional", func(t *testing.T) {
		attr, ok := attrs["vsock"]
		if !ok {
			t.Fatal("missing vsock attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("vsock must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("numa_optional", func(t *testing.T) {
		attr, ok := attrs["numa"]
		if !ok {
			t.Fatal("missing numa attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("numa must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("platform_optional", func(t *testing.T) {
		attr, ok := attrs["platform"]
		if !ok {
			t.Fatal("missing platform attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("platform must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("tpm_optional", func(t *testing.T) {
		attr, ok := attrs["tpm"]
		if !ok {
			t.Fatal("missing tpm attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("tpm must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("landlock_enable_optional", func(t *testing.T) {
		attr, ok := attrs["landlock_enable"]
		if !ok {
			t.Fatal("missing landlock_enable attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("landlock_enable must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("landlock_rules_optional", func(t *testing.T) {
		attr, ok := attrs["landlock_rules"]
		if !ok {
			t.Fatal("missing landlock_rules attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("landlock_rules must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("rate_limit_groups_optional", func(t *testing.T) {
		attr, ok := attrs["rate_limit_groups"]
		if !ok {
			t.Fatal("missing rate_limit_groups attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("rate_limit_groups must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("fs_optional", func(t *testing.T) {
		attr, ok := attrs["fs"]
		if !ok {
			t.Fatal("missing fs attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("fs must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("pmem_optional", func(t *testing.T) {
		attr, ok := attrs["pmem"]
		if !ok {
			t.Fatal("missing pmem attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("pmem must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("devices_optional", func(t *testing.T) {
		attr, ok := attrs["devices"]
		if !ok {
			t.Fatal("missing devices attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("devices must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("vdpa_optional", func(t *testing.T) {
		attr, ok := attrs["vdpa"]
		if !ok {
			t.Fatal("missing vdpa attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("vdpa must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("pci_segments_optional", func(t *testing.T) {
		attr, ok := attrs["pci_segments"]
		if !ok {
			t.Fatal("missing pci_segments attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("pci_segments must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("rtc_optional", func(t *testing.T) {
		attr, ok := attrs["rtc"]
		if !ok {
			t.Fatal("missing rtc attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("rtc must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("balloon_optional", func(t *testing.T) {
		attr, ok := attrs["balloon"]
		if !ok {
			t.Fatal("missing balloon attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("balloon must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("rng_optional", func(t *testing.T) {
		attr, ok := attrs["rng"]
		if !ok {
			t.Fatal("missing rng attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("rng must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("debug_console_optional", func(t *testing.T) {
		attr, ok := attrs["debug_console"]
		if !ok {
			t.Fatal("missing debug_console attribute")
		}
		_, ok = attr.(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("debug_console must be SingleNestedAttribute, got %T", attr)
		}
	})

	t.Run("user_devices_optional", func(t *testing.T) {
		attr, ok := attrs["user_devices"]
		if !ok {
			t.Fatal("missing user_devices attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("user_devices must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("generic_vhost_user_optional", func(t *testing.T) {
		attr, ok := attrs["generic_vhost_user"]
		if !ok {
			t.Fatal("missing generic_vhost_user attribute")
		}
		_, ok = attr.(schema.ListNestedAttribute)
		if !ok {
			t.Fatalf("generic_vhost_user must be ListNestedAttribute, got %T", attr)
		}
	})

	t.Run("pvpanic_optional", func(t *testing.T) {
		attr, ok := attrs["pvpanic"]
		if !ok {
			t.Fatal("missing pvpanic attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("pvpanic must be BoolAttribute, got %T", attr)
		}
	})
}

// ---------------------------------------------------------------------------
// Model Conversion Tests
// ---------------------------------------------------------------------------

func TestVMResourceModel_toClientConfig_Minimal(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{
			Kernel:  types.StringValue("/vmlinuz"),
			Cmdline: types.StringValue("console=hvc0"),
		},
	}

	cfg := model.toClientConfig()
	if cfg == nil {
		t.Fatal("toClientConfig returned nil")
	}
	if cfg.Payload == nil {
		t.Fatal("Payload is nil")
	}
	if cfg.Payload.Kernel != "/vmlinuz" {
		t.Errorf("Kernel = %q, want %q", cfg.Payload.Kernel, "/vmlinuz")
	}
	if cfg.Payload.Cmdline != "console=hvc0" {
		t.Errorf("Cmdline = %q, want %q", cfg.Payload.Cmdline, "console=hvc0")
	}
	if cfg.Cpus != nil {
		t.Error("Cpus must be nil for minimal config")
	}
	if cfg.Memory != nil {
		t.Error("Memory must be nil for minimal config")
	}
	if cfg.Disks != nil {
		t.Error("Disks must be nil for minimal config")
	}
}

func TestVMResourceModel_toClientConfig_FullPayload(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{
			Firmware:  types.StringValue("fw.bin"),
			Kernel:    types.StringValue("/vmlinuz"),
			Cmdline:   types.StringValue("console=hvc0"),
			Initramfs: types.StringValue("/initrd"),
			IGVM:      types.StringValue("boot.igvm"),
			HostData:  types.StringValue("base64data"),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Payload.Firmware != "fw.bin" {
		t.Errorf("Firmware = %q, want %q", cfg.Payload.Firmware, "fw.bin")
	}
	if cfg.Payload.Kernel != "/vmlinuz" {
		t.Errorf("Kernel = %q", cfg.Payload.Kernel)
	}
	if cfg.Payload.Initramfs != "/initrd" {
		t.Errorf("Initramfs = %q", cfg.Payload.Initramfs)
	}
	if cfg.Payload.IGVM != "boot.igvm" {
		t.Errorf("IGVM = %q", cfg.Payload.IGVM)
	}
	if cfg.Payload.HostData != "base64data" {
		t.Errorf("HostData = %q", cfg.Payload.HostData)
	}
}

func TestVMResourceModel_toClientConfig_WithCPU(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Cpus: &vmCpusModel{
			BootVcpus:   types.Int64Value(2),
			MaxVcpus:    types.Int64Value(4),
			KvmHyperv:   types.BoolValue(true),
			MaxPhysBits: types.Int64Value(40),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Cpus == nil {
		t.Fatal("Cpus is nil")
	}
	if cfg.Cpus.BootVcpus != 2 {
		t.Errorf("BootVcpus = %d, want 2", cfg.Cpus.BootVcpus)
	}
	if cfg.Cpus.MaxVcpus != 4 {
		t.Errorf("MaxVcpus = %d, want 4", cfg.Cpus.MaxVcpus)
	}
	if cfg.Cpus.KvmHyperv == nil || !*cfg.Cpus.KvmHyperv {
		t.Error("KvmHyperv should be true")
	}
	if cfg.Cpus.MaxPhysBits == nil || *cfg.Cpus.MaxPhysBits != 40 {
		t.Error("MaxPhysBits should be 40")
	}
}

func TestVMResourceModel_toClientConfig_WithCPUTopology(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Cpus: &vmCpusModel{
			BootVcpus: types.Int64Value(2),
			MaxVcpus:  types.Int64Value(4),
			Topology: &vmCpuTopologyModel{
				ThreadsPerCore: types.Int64Value(1),
				CoresPerDie:    types.Int64Value(2),
				DiesPerPackage: types.Int64Value(1),
				Packages:       types.Int64Value(1),
			},
		},
	}

	cfg := model.toClientConfig()
	if cfg.Cpus.Topology == nil {
		t.Fatal("Topology is nil")
	}
	if *cfg.Cpus.Topology.ThreadsPerCore != 1 {
		t.Errorf("ThreadsPerCore = %d", *cfg.Cpus.Topology.ThreadsPerCore)
	}
	if *cfg.Cpus.Topology.CoresPerDie != 2 {
		t.Errorf("CoresPerDie = %d", *cfg.Cpus.Topology.CoresPerDie)
	}
}

func TestVMResourceModel_toClientConfig_WithMemory(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Memory: &vmMemoryModel{
			Size:        types.Int64Value(2147483648),
			HotplugSize: types.Int64Value(4294967296),
			Mergeable:   types.BoolValue(true),
			Shared:      types.BoolValue(false),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Memory == nil {
		t.Fatal("Memory is nil")
	}
	if cfg.Memory.Size != 2147483648 {
		t.Errorf("Size = %d", cfg.Memory.Size)
	}
	if *cfg.Memory.HotplugSize != 4294967296 {
		t.Errorf("HotplugSize = %d", *cfg.Memory.HotplugSize)
	}
	if !*cfg.Memory.Mergeable {
		t.Error("Mergeable should be true")
	}
	if *cfg.Memory.Shared {
		t.Error("Shared should be false")
	}
}

func TestVMResourceModel_toClientConfig_WithSerial(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Serial: &vmSerialModel{
			File: types.StringValue("/tmp/serial.log"),
			Mode: types.StringValue("File"),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Serial == nil {
		t.Fatal("Serial is nil")
	}
	if cfg.Serial.File != "/tmp/serial.log" {
		t.Errorf("File = %q", cfg.Serial.File)
	}
	if string(cfg.Serial.Mode) != "File" {
		t.Errorf("Mode = %q", cfg.Serial.Mode)
	}
}

func TestVMResourceModel_toClientConfig_WithConsole(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Console: &vmConsoleModel{
			File:  types.StringValue("/tmp/console.log"),
			Mode:  types.StringValue("File"),
			Iommu: types.BoolValue(false),
			ID:    types.StringValue("console-0"),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Console == nil {
		t.Fatal("Console is nil")
	}
	if cfg.Console.File != "/tmp/console.log" {
		t.Errorf("File = %q", cfg.Console.File)
	}
	if cfg.Console.ID != "console-0" {
		t.Errorf("ID = %q", cfg.Console.ID)
	}
}

func TestVMResourceModel_toClientConfig_WithVsock(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Vsock: &vmVsockModel{
			CID:    types.Int64Value(3),
			Socket: types.StringValue("/tmp/vsock.sock"),
			Iommu:  types.BoolValue(true),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Vsock == nil {
		t.Fatal("Vsock is nil")
	}
	if cfg.Vsock.CID != 3 {
		t.Errorf("CID = %d", cfg.Vsock.CID)
	}
	if cfg.Vsock.Socket != "/tmp/vsock.sock" {
		t.Errorf("Socket = %q", cfg.Vsock.Socket)
	}
}

func TestVMResourceModel_toClientConfig_WithRng(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Rng: &vmRngModel{
			Src:   types.StringValue("/dev/urandom"),
			Iommu: types.BoolValue(false),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Rng == nil {
		t.Fatal("Rng is nil")
	}
	if cfg.Rng.Src != "/dev/urandom" {
		t.Errorf("Src = %q", cfg.Rng.Src)
	}
}

func TestVMResourceModel_toClientConfig_WithBalloon(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Balloon: &vmBalloonModel{
			Size:              types.Int64Value(1073741824),
			DeflateOnOOM:      types.BoolValue(true),
			FreePageReporting: types.BoolValue(true),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Balloon == nil {
		t.Fatal("Balloon is nil")
	}
	if cfg.Balloon.Size != 1073741824 {
		t.Errorf("Size = %d", cfg.Balloon.Size)
	}
}

func TestVMResourceModel_toClientConfig_WithDisks(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Disks: []vmDiskModel{
			{
				Path:     types.StringValue("/disk.img"),
				Readonly: types.BoolValue(false),
				Direct:   types.BoolValue(true),
				ID:       types.StringValue("rootfs"),
			},
			{
				Path:   types.StringValue("/disk2.img"),
				Sparse: types.BoolValue(true),
			},
		},
	}

	cfg := model.toClientConfig()
	if len(cfg.Disks) != 2 {
		t.Fatalf("Disks count = %d, want 2", len(cfg.Disks))
	}
	if cfg.Disks[0].Path != "/disk.img" {
		t.Errorf("Disks[0].Path = %q", cfg.Disks[0].Path)
	}
	if cfg.Disks[0].ID != "rootfs" {
		t.Errorf("Disks[0].ID = %q", cfg.Disks[0].ID)
	}
	if *cfg.Disks[1].Sparse != true {
		t.Error("Disks[1].Sparse should be true")
	}
}

func TestVMResourceModel_toClientConfig_WithNet(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Net: []vmNetModel{
			{
				Tap:       types.StringValue("chtap0"),
				MAC:       types.StringValue("de:ad:be:ef:00:01"),
				MTU:       types.Int64Value(1500),
				NumQueues: types.Int64Value(2),
			},
		},
	}

	cfg := model.toClientConfig()
	if len(cfg.Net) != 1 {
		t.Fatalf("Net count = %d, want 1", len(cfg.Net))
	}
	if cfg.Net[0].Tap != "chtap0" {
		t.Errorf("Net[0].Tap = %q", cfg.Net[0].Tap)
	}
	if cfg.Net[0].MAC != "de:ad:be:ef:00:01" {
		t.Errorf("Net[0].MAC = %q", cfg.Net[0].MAC)
	}
}

func TestVMResourceModel_toClientConfig_WithPlatform(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Platform: &vmPlatformModel{
			NumPCISegments: types.Int64Value(1),
			SystemUUID:     types.StringValue("550e8400-e29b-41d4-a716-446655440000"),
			Tdx:            types.BoolValue(true),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Platform == nil {
		t.Fatal("Platform is nil")
	}
	if *cfg.Platform.NumPCISegments != 1 {
		t.Errorf("NumPCISegments = %d", *cfg.Platform.NumPCISegments)
	}
	if cfg.Platform.SystemUUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("SystemUUID = %q", cfg.Platform.SystemUUID)
	}
}

func TestVMResourceModel_toClientConfig_WithTpm(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Tpm: &vmTpmModel{
			Socket: types.StringValue("/tmp/tpm.sock"),
		},
	}

	cfg := model.toClientConfig()
	if cfg.Tpm == nil {
		t.Fatal("Tpm is nil")
	}
	if cfg.Tpm.Socket != "/tmp/tpm.sock" {
		t.Errorf("Socket = %q", cfg.Tpm.Socket)
	}
}

func TestVMResourceModel_toClientConfig_WithFs(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		FS: []vmFsModel{
			{
				Tag:       types.StringValue("shared"),
				Socket:    types.StringValue("/tmp/fs.sock"),
				NumQueues: types.Int64Value(1),
				QueueSize: types.Int64Value(256),
			},
		},
	}

	cfg := model.toClientConfig()
	if len(cfg.FS) != 1 {
		t.Fatalf("FS count = %d", len(cfg.FS))
	}
	if cfg.FS[0].Tag != "shared" {
		t.Errorf("Tag = %q", cfg.FS[0].Tag)
	}
}

func TestVMResourceModel_toClientConfig_WithPmem(t *testing.T) {
	model := &vmResourceModel{
		Payload: &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Pmem: []vmPmemModel{
			{
				File: types.StringValue("/nvdimm.img"),
				Size: types.Int64Value(268435456),
			},
		},
	}

	cfg := model.toClientConfig()
	if len(cfg.Pmem) != 1 {
		t.Fatalf("Pmem count = %d", len(cfg.Pmem))
	}
	if cfg.Pmem[0].File != "/nvdimm.img" {
		t.Errorf("File = %q", cfg.Pmem[0].File)
	}
}

func TestVMResourceModel_toClientConfig_NullFieldsAreNil(t *testing.T) {
	model := &vmResourceModel{
		Payload:  &vmPayloadModel{Kernel: types.StringValue("/vmlinuz")},
		Watchdog: types.BoolValue(true),
		Iommu:    types.BoolValue(false),
	}

	cfg := model.toClientConfig()
	if *cfg.Watchdog != true {
		t.Error("Watchdog should be true")
	}
	if *cfg.Iommu != false {
		t.Error("Iommu should be false")
	}
}

// ---------------------------------------------------------------------------
// Resource Metadata Test
// ---------------------------------------------------------------------------

func TestVMResource_Metadata(t *testing.T) {
	r := newVMResource()
	var resp fwresource.MetadataResponse
	r.Metadata(context.Background(), fwresource.MetadataRequest{}, &resp)
	if resp.TypeName != "cloudhypervisor_vm" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "cloudhypervisor_vm")
	}
}

// ---------------------------------------------------------------------------
// Test Helpers (used by other resource tests)
// ---------------------------------------------------------------------------

// testAccExternalCHHelper starts a cloud-hypervisor process for
// acceptance tests, returning the socket path and a cleanup function.
func testAccExternalCHHelper(t *testing.T) (socketPath string, cleanup func()) {
	t.Helper()

	fakeBin, err := exec.LookPath("cloud-hypervisor")
	if err != nil {
		t.Skipf("cloud-hypervisor binary not found on PATH: %v", err)
	}

	socketDir := t.TempDir()
	socketPath = filepath.Join(socketDir, "api.sock")

	cmd := exec.Command(fakeBin, "--api-socket-path", socketPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start cloud-hypervisor: %v", err)
	}

	go func() {
		cmd.Wait()
	}()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", socketPath, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	cleanup = func() {
		_ = syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(500 * time.Millisecond)
		_ = syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		_ = os.RemoveAll(socketDir)
	}

	return socketPath, cleanup
}

// testAccCHProcessManager creates a chproc.Manager with the given binary
// path and starts it, returning the socket path and stop function.
func testAccCHProcessManager(t *testing.T, binaryPath string) (socketPath string, mgr *chproc.Manager, stop func()) {
	t.Helper()

	mgr = chproc.NewManager(
		chproc.WithBinaryPath(binaryPath),
	)

	ctx := context.Background()
	var err error
	socketPath, err = mgr.Start(ctx)
	if err != nil {
		t.Fatalf("chproc.Manager.Start: %v", err)
	}

	if err := mgr.WaitReady(ctx, 30*time.Second); err != nil {
		t.Fatalf("WaitReady: %v", err)
	}

	stop = func() {
		_ = mgr.Stop(context.Background())
	}

	return socketPath, mgr, stop
}

// findFreeTCPPort finds a free TCP port for testing.
func findFreeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free TCP port: %v", err)
	}
	defer listener.Close()
	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("failed to convert port: %v", err)
	}
	return port
}

// testAccVMResourceBasicConfig returns an HCL config for a minimal VM.
func testAccVMResourceBasicConfig(kernelPath, initrdPath string) string {
	return fmt.Sprintf(`
resource "cloudhypervisor_vm" "test" {
  payload = {
    kernel    = %q
    initramfs = %q
    cmdline   = "console=ttyS0"
  }
}
`, kernelPath, initrdPath)
}

// testAccVMResourceCheckDestroy verifies the VM resource no longer exists.
func testAccVMResourceCheckDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudhypervisor_vm" {
			continue
		}
		return fmt.Errorf("resource %s still exists", rs.Primary.ID)
	}
	return nil
}

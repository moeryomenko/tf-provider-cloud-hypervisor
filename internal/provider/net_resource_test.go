package provider

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Schema Validation Tests
// ---------------------------------------------------------------------------

// TestNetResourceSchema verifies all top-level net resource attributes.
func TestNetResourceSchema(t *testing.T) {
	r := newNetResource()
	var resp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes

	t.Run("vm_socket_path_required", func(t *testing.T) {
		attr, ok := attrs["vm_socket_path"]
		if !ok {
			t.Fatal("missing vm_socket_path attribute")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("vm_socket_path must be StringAttribute, got %T", attr)
		}
		if !strAttr.Required {
			t.Error("vm_socket_path must be Required")
		}
	})

	t.Run("tap_optional", func(t *testing.T) {
		attr, ok := attrs["tap"]
		if !ok {
			t.Fatal("missing tap attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("tap must be StringAttribute, got %T", attr)
		}
	})

	t.Run("mac_optional", func(t *testing.T) {
		attr, ok := attrs["mac"]
		if !ok {
			t.Fatal("missing mac attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("mac must be StringAttribute, got %T", attr)
		}
	})

	t.Run("mtu_optional", func(t *testing.T) {
		attr, ok := attrs["mtu"]
		if !ok {
			t.Fatal("missing mtu attribute")
		}
		_, ok = attr.(schema.Int64Attribute)
		if !ok {
			t.Fatalf("mtu must be Int64Attribute, got %T", attr)
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

	t.Run("num_queues_optional", func(t *testing.T) {
		attr, ok := attrs["num_queues"]
		if !ok {
			t.Fatal("missing num_queues attribute")
		}
		_, ok = attr.(schema.Int64Attribute)
		if !ok {
			t.Fatalf("num_queues must be Int64Attribute, got %T", attr)
		}
	})

	t.Run("vhost_user_optional", func(t *testing.T) {
		attr, ok := attrs["vhost_user"]
		if !ok {
			t.Fatal("missing vhost_user attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("vhost_user must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("vhost_mode_optional", func(t *testing.T) {
		attr, ok := attrs["vhost_mode"]
		if !ok {
			t.Fatal("missing vhost_mode attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("vhost_mode must be StringAttribute, got %T", attr)
		}
	})

	t.Run("device_id_computed", func(t *testing.T) {
		attr, ok := attrs["device_id"]
		if !ok {
			t.Fatal("missing device_id attribute")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("device_id must be StringAttribute, got %T", attr)
		}
		if !strAttr.Computed {
			t.Error("device_id must be Computed")
		}
	})

	t.Run("bdf_computed", func(t *testing.T) {
		attr, ok := attrs["bdf"]
		if !ok {
			t.Fatal("missing bdf attribute")
		}
		strAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("bdf must be StringAttribute, got %T", attr)
		}
		if !strAttr.Computed {
			t.Error("bdf must be Computed")
		}
	})

	t.Run("ip_optional", func(t *testing.T) {
		attr, ok := attrs["ip"]
		if !ok {
			t.Fatal("missing ip attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("ip must be StringAttribute, got %T", attr)
		}
	})

	t.Run("mask_optional", func(t *testing.T) {
		attr, ok := attrs["mask"]
		if !ok {
			t.Fatal("missing mask attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("mask must be StringAttribute, got %T", attr)
		}
	})

	t.Run("host_mac_optional", func(t *testing.T) {
		attr, ok := attrs["host_mac"]
		if !ok {
			t.Fatal("missing host_mac attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("host_mac must be StringAttribute, got %T", attr)
		}
	})

	t.Run("queue_size_optional", func(t *testing.T) {
		attr, ok := attrs["queue_size"]
		if !ok {
			t.Fatal("missing queue_size attribute")
		}
		_, ok = attr.(schema.Int64Attribute)
		if !ok {
			t.Fatalf("queue_size must be Int64Attribute, got %T", attr)
		}
	})

	t.Run("vhost_socket_optional", func(t *testing.T) {
		attr, ok := attrs["vhost_socket"]
		if !ok {
			t.Fatal("missing vhost_socket attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("vhost_socket must be StringAttribute, got %T", attr)
		}
	})

	t.Run("offload_tso_optional", func(t *testing.T) {
		attr, ok := attrs["offload_tso"]
		if !ok {
			t.Fatal("missing offload_tso attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("offload_tso must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("offload_ufo_optional", func(t *testing.T) {
		attr, ok := attrs["offload_ufo"]
		if !ok {
			t.Fatal("missing offload_ufo attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("offload_ufo must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("offload_csum_optional", func(t *testing.T) {
		attr, ok := attrs["offload_csum"]
		if !ok {
			t.Fatal("missing offload_csum attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("offload_csum must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("pci_segment_optional", func(t *testing.T) {
		attr, ok := attrs["pci_segment"]
		if !ok {
			t.Fatal("missing pci_segment attribute")
		}
		_, ok = attr.(schema.Int64Attribute)
		if !ok {
			t.Fatalf("pci_segment must be Int64Attribute, got %T", attr)
		}
	})

	t.Run("pci_device_id_optional", func(t *testing.T) {
		attr, ok := attrs["pci_device_id"]
		if !ok {
			t.Fatal("missing pci_device_id attribute")
		}
		_, ok = attr.(schema.Int64Attribute)
		if !ok {
			t.Fatalf("pci_device_id must be Int64Attribute, got %T", attr)
		}
	})
}

// ---------------------------------------------------------------------------
// Model Conversion Tests
// ---------------------------------------------------------------------------

func TestNetResourceModel_toClientConfig_Minimal(t *testing.T) {
	model := &netResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Tap:          types.StringValue("chtap0"),
	}

	cfg := model.toClientConfig()
	if cfg == nil {
		t.Fatal("toClientConfig returned nil")
	}
	if cfg.Tap != "chtap0" {
		t.Errorf("Tap = %q, want %q", cfg.Tap, "chtap0")
	}
	if cfg.MAC != "" {
		t.Errorf("MAC = %q, want empty", cfg.MAC)
	}
	if cfg.MTU != nil {
		t.Error("MTU must be nil when not set")
	}
}

func TestNetResourceModel_toClientConfig_Full(t *testing.T) {
	model := &netResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Tap:          types.StringValue("chtap0"),
		IP:           types.StringValue("192.168.1.100"),
		Mask:         types.StringValue("255.255.255.0"),
		MAC:          types.StringValue("de:ad:be:ef:00:01"),
		HostMAC:      types.StringValue("de:ad:be:ef:00:02"),
		MTU:          types.Int64Value(1500),
		Iommu:        types.BoolValue(false),
		NumQueues:    types.Int64Value(2),
		QueueSize:    types.Int64Value(256),
		VhostUser:    types.BoolValue(false),
		VhostSocket:  types.StringValue("/tmp/vhost.sock"),
		VhostMode:    types.StringValue("client"),
		ID:           types.StringValue("mynet"),
		PCISegment:   types.Int64Value(0),
		PCIDeviceID:  types.Int64Value(1),
		OffloadTSO:   types.BoolValue(true),
		OffloadUFO:   types.BoolValue(false),
		OffloadCsum:  types.BoolValue(true),
	}

	cfg := model.toClientConfig()
	if cfg.Tap != "chtap0" {
		t.Errorf("Tap = %q", cfg.Tap)
	}
	if cfg.IP != "192.168.1.100" {
		t.Errorf("IP = %q", cfg.IP)
	}
	if cfg.Mask != "255.255.255.0" {
		t.Errorf("Mask = %q", cfg.Mask)
	}
	if cfg.MAC != "de:ad:be:ef:00:01" {
		t.Errorf("MAC = %q", cfg.MAC)
	}
	if cfg.HostMAC != "de:ad:be:ef:00:02" {
		t.Errorf("HostMAC = %q", cfg.HostMAC)
	}
	if *cfg.MTU != 1500 {
		t.Errorf("MTU = %d", *cfg.MTU)
	}
	if *cfg.Iommu {
		t.Error("Iommu should be false")
	}
	if *cfg.NumQueues != 2 {
		t.Errorf("NumQueues = %d", *cfg.NumQueues)
	}
	if *cfg.QueueSize != 256 {
		t.Errorf("QueueSize = %d", *cfg.QueueSize)
	}
	if *cfg.VhostUser {
		t.Error("VhostUser should be false")
	}
	if cfg.VhostSocket != "/tmp/vhost.sock" {
		t.Errorf("VhostSocket = %q", cfg.VhostSocket)
	}
	if cfg.VhostMode != "client" {
		t.Errorf("VhostMode = %q", cfg.VhostMode)
	}
	if cfg.ID != "mynet" {
		t.Errorf("ID = %q", cfg.ID)
	}
	if *cfg.OffloadTSO != true {
		t.Error("OffloadTSO should be true")
	}
	if *cfg.OffloadUFO != false {
		t.Error("OffloadUFO should be false")
	}
	if *cfg.OffloadCsum != true {
		t.Error("OffloadCsum should be true")
	}
}

func TestNetResourceModel_toClientConfig_NullOptionalFields(t *testing.T) {
	model := &netResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Tap:          types.StringValue("chtap0"),
	}

	cfg := model.toClientConfig()
	if cfg.IP != "" {
		t.Error("IP must be empty when not set")
	}
	if cfg.MTU != nil {
		t.Error("MTU must be nil when not set")
	}
	if cfg.NumQueues != nil {
		t.Error("NumQueues must be nil when not set")
	}
	if cfg.VhostUser != nil {
		t.Error("VhostUser must be nil when not set")
	}
	if cfg.OffloadTSO != nil {
		t.Error("OffloadTSO must be nil when not set")
	}
	if cfg.PCIDeviceID != nil {
		t.Error("PCIDeviceID must be nil when not set")
	}
}

func TestNetResourceModel_toClientConfig_BoolValues(t *testing.T) {
	model := &netResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Tap:          types.StringValue("chtap0"),
		Iommu:        types.BoolValue(true),
		VhostUser:    types.BoolValue(true),
		OffloadTSO:   types.BoolValue(true),
		OffloadCsum:  types.BoolValue(true),
	}

	cfg := model.toClientConfig()
	if !*cfg.Iommu {
		t.Error("Iommu should be true")
	}
	if !*cfg.VhostUser {
		t.Error("VhostUser should be true")
	}
	if !*cfg.OffloadTSO {
		t.Error("OffloadTSO should be true")
	}
}

// ---------------------------------------------------------------------------
// Resource Metadata Test
// ---------------------------------------------------------------------------

func TestNetResource_Metadata(t *testing.T) {
	r := newNetResource()
	var resp fwresource.MetadataResponse
	r.Metadata(context.Background(), fwresource.MetadataRequest{}, &resp)
	if resp.TypeName != "cloudhypervisor_net" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "cloudhypervisor_net")
	}
}

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

// TestDiskResourceSchema verifies all top-level disk resource attributes.
func TestDiskResourceSchema(t *testing.T) {
	r := newDiskResource()
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

	t.Run("path_optional", func(t *testing.T) {
		attr, ok := attrs["path"]
		if !ok {
			t.Fatal("missing path attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("path must be StringAttribute, got %T", attr)
		}
	})

	t.Run("readonly_optional", func(t *testing.T) {
		attr, ok := attrs["readonly"]
		if !ok {
			t.Fatal("missing readonly attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("readonly must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("direct_optional", func(t *testing.T) {
		attr, ok := attrs["direct"]
		if !ok {
			t.Fatal("missing direct attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("direct must be BoolAttribute, got %T", attr)
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

	t.Run("id_optional", func(t *testing.T) {
		attr, ok := attrs["id"]
		if !ok {
			t.Fatal("missing id attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("id must be StringAttribute, got %T", attr)
		}
	})

	t.Run("serial_optional", func(t *testing.T) {
		attr, ok := attrs["serial"]
		if !ok {
			t.Fatal("missing serial attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("serial must be StringAttribute, got %T", attr)
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

	t.Run("sparse_optional", func(t *testing.T) {
		attr, ok := attrs["sparse"]
		if !ok {
			t.Fatal("missing sparse attribute")
		}
		_, ok = attr.(schema.BoolAttribute)
		if !ok {
			t.Fatalf("sparse must be BoolAttribute, got %T", attr)
		}
	})

	t.Run("rate_limit_group_optional", func(t *testing.T) {
		attr, ok := attrs["rate_limit_group"]
		if !ok {
			t.Fatal("missing rate_limit_group attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("rate_limit_group must be StringAttribute, got %T", attr)
		}
	})

	t.Run("image_type_optional", func(t *testing.T) {
		attr, ok := attrs["image_type"]
		if !ok {
			t.Fatal("missing image_type attribute")
		}
		_, ok = attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("image_type must be StringAttribute, got %T", attr)
		}
	})
}

// ---------------------------------------------------------------------------
// Model Conversion Tests
// ---------------------------------------------------------------------------

func TestDiskResourceModel_toClientConfig_Minimal(t *testing.T) {
	model := &diskResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Path:         types.StringValue("/disk.img"),
	}

	cfg := model.toClientConfig()
	if cfg == nil {
		t.Fatal("toClientConfig returned nil")
	}
	if cfg.Path != "/disk.img" {
		t.Errorf("Path = %q, want %q", cfg.Path, "/disk.img")
	}
	if cfg.Readonly != nil {
		t.Error("Readonly must be nil when not set")
	}
}

func TestDiskResourceModel_toClientConfig_Full(t *testing.T) {
	model := &diskResourceModel{
		VMSocketPath:    types.StringValue("/tmp/sock"),
		Path:            types.StringValue("/disk.img"),
		Readonly:        types.BoolValue(true),
		Direct:          types.BoolValue(true),
		Iommu:           types.BoolValue(false),
		NumQueues:       types.Int64Value(2),
		QueueSize:       types.Int64Value(256),
		VhostUser:       types.BoolValue(false),
		VhostSocket:     types.StringValue("/tmp/vhost.sock"),
		PCISegment:      types.Int64Value(0),
		PCIDeviceID:     types.Int64Value(1),
		ID:              types.StringValue("mydisk"),
		Serial:          types.StringValue("DISK-123"),
		RateLimitGroup:  types.StringValue("rl-group"),
		BackingFiles:    types.BoolValue(true),
		Sparse:          types.BoolValue(true),
		ImageType:       types.StringValue("raw"),
		LockGranularity: types.StringValue("sector"),
	}

	cfg := model.toClientConfig()
	if cfg.Path != "/disk.img" {
		t.Errorf("Path = %q", cfg.Path)
	}
	if !*cfg.Readonly {
		t.Error("Readonly should be true")
	}
	if !*cfg.Direct {
		t.Error("Direct should be true")
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
	if *cfg.PCIDeviceID != 1 {
		t.Errorf("PCIDeviceID = %d", *cfg.PCIDeviceID)
	}
	if cfg.ID != "mydisk" {
		t.Errorf("ID = %q", cfg.ID)
	}
	if cfg.Serial != "DISK-123" {
		t.Errorf("Serial = %q", cfg.Serial)
	}
	if *cfg.BackingFiles != true {
		t.Error("BackingFiles should be true")
	}
	if *cfg.Sparse != true {
		t.Error("Sparse should be true")
	}
}

func TestDiskResourceModel_toClientConfig_NullOptionalFields(t *testing.T) {
	model := &diskResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Path:         types.StringValue("/disk.img"),
	}

	cfg := model.toClientConfig()
	if cfg.NumQueues != nil {
		t.Error("NumQueues must be nil when not set")
	}
	if cfg.VhostUser != nil {
		t.Error("VhostUser must be nil when not set")
	}
	if cfg.PCIDeviceID != nil {
		t.Error("PCIDeviceID must be nil when not set")
	}
	if cfg.Sparse != nil {
		t.Error("Sparse must be nil when not set")
	}
}

func TestDiskResourceModel_toClientConfig_WithRateLimiter(t *testing.T) {
	model := &diskResourceModel{
		VMSocketPath: types.StringValue("/tmp/sock"),
		Path:         types.StringValue("/disk.img"),
	}

	cfg := model.toClientConfig()
	if cfg.RateLimiterConfig != nil {
		t.Error("RateLimiterConfig must be nil when not set")
	}
}

// ---------------------------------------------------------------------------
// Resource Metadata Test
// ---------------------------------------------------------------------------

func TestDiskResource_Metadata(t *testing.T) {
	r := newDiskResource()
	var resp fwresource.MetadataResponse
	r.Metadata(context.Background(), fwresource.MetadataRequest{}, &resp)
	if resp.TypeName != "cloudhypervisor_disk" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "cloudhypervisor_disk")
	}
}

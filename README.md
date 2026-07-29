# Terraform Provider for Cloud-Hypervisor

This provider allows managing [Cloud-Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor) virtual machines using Terraform or OpenTofu. It communicates with Cloud-Hypervisor's REST API to define, boot, and manage VMs and their devices.

## Overview

Cloud-Hypervisor is a lightweight VMM built on KVM. This provider wraps its HTTP API to manage the full VM lifecycle — creation, boot, shutdown, deletion — and supports hotplugging disks and network devices into running VMs.

The provider can either manage the Cloud-Hypervisor process automatically (starting and stopping it alongside resources) or connect to an externally-managed Cloud-Hypervisor instance.

## Resources

| Resource | Description |
|----------|-------------|
| `cloudhypervisor_vm` | Full VM lifecycle: create, boot, shutdown, delete. Supports kernel/initrd or firmware payload, CPU topology, memory configuration, inline disks and network devices, virtio-fs, PMEM, VFIO passthrough, vDPA, vsock, RNG, balloon, RTC, TPM, NUMA, PCI segments, and platform configuration. |
| `cloudhypervisor_disk` | Hotplug a disk into a running VM. References the parent VM via its `socket_path`. |
| `cloudhypervisor_net` | Hotplug a network device into a running VM. References the parent VM via its `socket_path`. |

## Using the Provider

```hcl
terraform {
  required_providers {
    cloudhypervisor = {
      source = "registry.terraform.io/community/cloudhypervisor"
    }
  }
}

provider "cloudhypervisor" {
  # Defaults: manage_ch_process = true, ch_binary_path = "cloud-hypervisor"
  # The provider will start cloud-hypervisor automatically.
}

resource "cloudhypervisor_vm" "example" {
  payload = {
    kernel    = "/path/to/vmlinux"
    initramfs = "/path/to/initrd.img"
    cmdline   = "console=ttyS0 root=/dev/vda1 rw"
  }

  cpus = {
    boot_vcpus = 2
    max_vcpus  = 4
  }

  memory = {
    size = 2147483648  # 2 GiB
  }
}
```

### Connection Modes

The provider supports two connection modes:

**Managed mode (default)** — The provider starts Cloud-Hypervisor automatically, managing its full lifecycle. Communication happens over a Unix domain socket.

```hcl
provider "cloudhypervisor" {
  # manage_ch_process defaults to true
  # ch_binary_path defaults to "cloud-hypervisor"
}
```

**External mode** — Connect to an already-running Cloud-Hypervisor instance via its HTTP API.

```hcl
provider "cloudhypervisor" {
  manage_ch_process = false
  ch_http_api       = "http://192.168.1.100:8080/api/v1"
}
```

See the [examples](./examples) directory for complete usage:
- [vm-basic](./examples/vm-basic/) — Minimal VM with kernel+initrd
- [vm-cloudinit](./examples/vm-cloudinit/) — VM booting an Ubuntu cloud image via firmware with cloud-init
- [vm-full](./examples/vm-full/) — Full-featured VM with all supported device types
- [vm-multi-disk](./examples/vm-multi-disk/) — VM with separate OS and data disks
- [hotplug](./examples/hotplug/) — VM with hotplugged disk and network sub-resources

## Design

- **API Fidelity** — The Terraform schema maps directly to the Cloud-Hypervisor REST API's `VmConfig` structure, giving users full access to Cloud-Hypervisor's features.
- **Plugin Framework** — Built with Terraform Plugin Framework (not legacy SDK v2).
- **Managed Process Lifecycle** — In managed mode, the provider creates a temporary socket directory, starts Cloud-Hypervisor with API access on a Unix domain socket, and cleans up the process on destroy.
- **Sub-Resource Hotplug** — Disk and network sub-resources use the parent VM's `socket_path` to find and attach to the running Cloud-Hypervisor instance for hotplug operations.

## Development

### Prerequisites

- Go 1.26+
- [Cloud-Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor) binary (for tests and managed mode)
- Linux with KVM (`/dev/kvm` accessible)

### Building

```bash
git clone https://github.com/moeryomenko/tf-provider-cloud-hypervisor
cd tf-provider-cloud-hypervisor
make build
```

### Installing locally

```bash
make install  # Optional: installs to ~/.terraform.d/plugins/
```

To use the locally-built provider, add a dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/community/cloudhypervisor" = "/path/to/tf-provider-cloud-hypervisor"
  }
  direct {}
}
```

### Running tests

```bash
# Lint
make lint

# Unit tests
make test

# Download acceptance test dependencies (kernel + initrd)
make testdeps-acc

# Acceptance tests (require KVM and Cloud-Hypervisor binary)
make testacc
```

### Available targets

```
build          Build the provider binary
test           Run all unit tests
testacc        Run acceptance tests (requires Cloud-Hypervisor)
lint           Run golangci-lint
fmt            Format Go source files
tidy           Tidy Go module dependencies
vet            Run go vet
clean          Clean build artifacts
testdeps-acc   Download kernel+initrd fixtures for acceptance tests
sweep          Clean up leaked test resources
check          Run lint, vet, and tests (CI gate)
```

## License

Licensed under either of:

- Apache License, Version 2.0 ([LICENSE-APACHE](./LICENSE-APACHE))
- MIT license ([LICENSE-MIT](./LICENSE-MIT))

at your option.

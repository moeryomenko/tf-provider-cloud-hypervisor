terraform {
  required_providers {
    cloudhypervisor = {
      source = "registry.terraform.io/community/cloudhypervisor"
    }
  }
}

provider "cloudhypervisor" {
  # manage_ch_process defaults to true; the provider starts cloud-hypervisor
  # automatically. To use an externally-managed instance, set
  #   manage_ch_process = false
  #   ch_http_api       = "http://host:port/api/v1"
}

# -----------------------------------------------------------------------------
# Full-featured VM example.
#
# NOTE: Inline disks and net blocks are create-time only. After the VM is
# created, use cloudhypervisor_disk and cloudhypervisor_net sub-resources
# for hotplug operations.
# -----------------------------------------------------------------------------
resource "cloudhypervisor_vm" "full" {
  payload = {
    kernel   = var.kernel_path
    initramfs = var.initrd_path
    cmdline  = "console=ttyS0 root=/dev/vda1 rw"
  }

  cpus = {
    boot_vcpus = 2
    max_vcpus  = 4
    topology = {
      threads_per_core = 1
      cores_per_die    = 2
      dies_per_package = 1
      packages         = 1
    }
    kvm_hyperv  = true
    max_phys_bits = 40
  }

  memory = {
    size         = 2147483648  # 2 GiB
    hotplug_size = 4294967296  # 4 GiB max
  }

  # Serial console output to a file
  serial = {
    file = "/tmp/ch-serial.log"
    mode = "File"
  }

  # Console device configuration
  console = {
    file  = "/tmp/ch-console.log"
    mode  = "File"
    iommu = false
  }

  # Random number generator
  rng = {
    src   = "/dev/urandom"
    iommu = false
  }

  # Balloon device for memory management
  balloon = {
    size               = 1073741824  # 1 GiB balloon
    deflate_on_oom     = true
    free_page_reporting = true
  }

  # Real-time clock
  rtc = {
    iommu = false
  }

  # Platform configuration
  platform = {
    num_pci_segments         = 1
    iommu_segments           = [0]
    iommu_address_width_bits = 48
    serial_number            = "CH-VM-001"
    uuid                     = "550e8400-e29b-41d4-a716-446655440000"
  }

  # TPM emulation socket
  tpm = {
    socket = "/tmp/ch-tpm.sock"
  }

  # vhost-user vsock
  vsock = {
    cid    = 3
    socket = "/tmp/ch-vsock.sock"
    iommu  = false
  }

  # Global IOMMU flag
  iommu = false

  # Watchdog device
  watchdog = true

  # Landlock sandboxing
  landlock_enable = false

  # Create-time inline disks (use cloudhypervisor_disk for hotplug)
  disks = [
    {
      path     = "/var/lib/ch/vms/full/rootfs.img"
      readonly = false
      direct   = true
      iommu    = false
      num_queues = 1
      id       = "rootfs"
      sparse   = false
    },
  ]

  # Create-time inline network (use cloudhypervisor_net for hotplug)
  net = [
    {
      tap        = "chtap0"
      mac        = "de:ad:be:ef:00:01"
      mtu        = 1500
      iommu      = false
      num_queues = 2
      vhost_user = false
    },
  ]

  # virtio-fs shared directory
  fs = [
    {
      tag    = "shared"
      socket = "/tmp/ch-fs.sock"
    },
  ]

  # Persistent memory device
  pmem = [
    {
      file = "/var/lib/ch/vms/full/nvdimm.img"
    },
  ]

  # VFIO device passthrough
  devices = [
    {
      path = "/sys/bus/pci/devices/0000:00:1f.0"
    },
  ]

  # vDPA device
  vdpa = [
    {
      path = "/dev/vdpa/vdpa0"
    },
  ]

  # Guest NUMA topology
  numa = [
    {
      guest_numa_id = 0
    },
  ]

  # PCI segment configuration
  pci_segments = [
    {
      pci_segment = 0
    },
  ]
}

variable "kernel_path" {
  description = "Path to Linux kernel binary"
  type = string
}

variable "initrd_path" {
  description = "Path to initrd image"
  type = string
}

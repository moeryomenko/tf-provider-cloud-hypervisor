# ------------------------------------------------------------------------------
# vm-multi-disk — VM with separate OS and data disks.
#
# This example creates a cloud-hypervisor VM with two virtio-blk disks:
#   - /dev/vda (OS disk, index 0, id = "os")
#   - /dev/vdb (data disk, index 1, id = "data")
#
# The data disk persists across VM restarts and must be formatted inside the
# guest (e.g., mkfs.ext4 /dev/vdb) before first use.
#
# Cloud-hypervisor assigns PCI device slots in array order, so the OS disk
# at index 0 always appears as the first block device in the guest.
#
# NOTE: Inline disks are create-time only. After the VM is created, use
# cloudhypervisor_disk and cloudhypervisor_net sub-resources for hotplug.
# ------------------------------------------------------------------------------

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

resource "cloudhypervisor_vm" "multi_disk_vm" {
  payload = {
    kernel   = var.kernel_path
    initramfs = var.initrd_path
    cmdline  = var.kernel_cmdline
  }

  cpus = {
    boot_vcpus = var.vcpus
    max_vcpus  = var.max_vcpus
  }

  memory = {
    size = var.memory_size_bytes  # bytes (not MiB)
  }

  # OS disk — appears as /dev/vda in the guest
  # Data disk — appears as /dev/vdb in the guest
  disks = [
    {
      path     = var.os_disk_path
      readonly = false
      direct   = true
      id       = "os"
    },
    {
      path     = var.data_disk_path
      readonly = false
      direct   = true
      id       = "data"
    },
  ]

  net = [
    {
      tap        = var.tap_device_name
      mac        = var.guest_mac_address
      mtu        = 1500
      num_queues = 2
    },
  ]

  rng = {
    src = "/dev/urandom"
  }

  serial = {
    file = var.serial_log_path
    mode = "File"
  }
}

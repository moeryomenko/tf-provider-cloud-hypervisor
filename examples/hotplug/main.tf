terraform {
  required_providers {
    cloudhypervisor = {
      source = "registry.terraform.io/community/cloudhypervisor"
    }
  }
}

provider "cloudhypervisor" {
  # manage_ch_process defaults to true; the provider starts cloud-hypervisor
  # automatically.
}

# -----------------------------------------------------------------------------
# Minimal VM — just enough to boot, using the provider-managed process.
# Sub-resources (disk, net) attach to this VM via socket_path.
# -----------------------------------------------------------------------------
resource "cloudhypervisor_vm" "demo" {
  payload = {
    kernel   = var.kernel_path
    initramfs = var.initrd_path
    cmdline  = "console=ttyS0 root=/dev/vda1 rw"
  }

  cpus = {
    boot_vcpus = 1
    max_vcpus  = 2
  }

  memory = {
    size = 1073741824  # 1 GiB
  }
}

# -----------------------------------------------------------------------------
# Hotplug a disk into the demo VM.
#
# The vm_socket_path attribute links this sub-resource to the parent VM.
# When the VM is destroyed, the socket path becomes unreachable and the disk
# resource is automatically removed from state on the next refresh.
# -----------------------------------------------------------------------------
resource "cloudhypervisor_disk" "extra" {
  vm_socket_path = cloudhypervisor_vm.demo.socket_path
  path           = "/var/lib/ch/disks/extra.img"
  readonly       = false
  direct         = true
  num_queues     = 1
}

# -----------------------------------------------------------------------------
# Hotplug a network device into the demo VM.
#
# Like the disk, the net sub-resource references the parent VM via
# vm_socket_path. The net device is hotplugged and can be removed without
# destroying the VM.
# -----------------------------------------------------------------------------
resource "cloudhypervisor_net" "extra" {
  vm_socket_path = cloudhypervisor_vm.demo.socket_path
  tap            = "chtap1"
  mac            = "de:ad:be:ef:00:02"
  mtu            = 1500
  num_queues     = 2
}

variable "kernel_path" {
  description = "Path to Linux kernel binary"
  type = string
}

variable "initrd_path" {
  description = "Path to initrd image"
  type = string
}

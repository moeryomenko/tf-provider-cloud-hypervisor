terraform {
  required_providers {
    cloudhypervisor = {
      source = "registry.terraform.io/community/cloudhypervisor"
    }
  }
}

provider "cloudhypervisor" {
  # Defaults: manage_ch_process = true, ch_binary_path = "cloud-hypervisor"
  # The provider will start cloud-hypervisor automatically
}

# Find your kernel + initrd paths at:
# CH_TEST_KERNEL=/path/to/vmlinux CH_TEST_INITRD=/path/to/initrd
resource "cloudhypervisor_vm" "basic" {
  payload = {
    kernel   = var.kernel_path
    initramfs = var.initrd_path
    cmdline  = "console=ttyS0 root=/dev/vda1 rw"
  }
}

variable "kernel_path" {
  description = "Path to Linux kernel binary"
  type = string
}

variable "initrd_path" {
  description = "Path to initrd image"
  type = string
}

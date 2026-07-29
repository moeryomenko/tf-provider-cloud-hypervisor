# -----------------------------------------------------------------------------
# VM with Cloud-Init Provisioning
#
# This example demonstrates booting an Ubuntu cloud image using firmware
# (hypervisor-fw or CLOUDHV.fd) with cloud-init for automated provisioning.
#
# Prerequisites:
#   1. Download and prepare an Ubuntu cloud image:
#        wget https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img
#        qemu-img convert -f qcow2 -O raw noble-server-cloudimg-amd64.img /path/to/ubuntu.raw
#
#   2. Download firmware:
#        wget https://github.com/cloud-hypervisor/rust-hypervisor-firmware/releases/.../hypervisor-fw
#
#   3. Create the cloud-init disk image:
#        ./create-cloud-init.sh
#
#   4. Create a tap device:
#        sudo ip tuntap add chtap0 mode tap
#        sudo ip link set chtap0 up
#
#   5. Apply this configuration:
#        tofu init && tofu plan -var-file=terraform.tfvars
#        tofu apply -var-file=terraform.tfvars
#
# NOTE: All file paths use var.* references for flexibility. See variables.tf
# for configurable options and terraform.tfvars.example for sample values.
# -----------------------------------------------------------------------------

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

resource "cloudhypervisor_vm" "cloud_vm" {
  payload = {
    firmware = var.firmware_path
    cmdline  = "console=hvc0 root=/dev/vda1 rw"
  }

  cpus = {
    boot_vcpus = var.vcpus
    max_vcpus  = var.max_vcpus
  }

  memory = {
    size = var.memory_size_bytes
  }

  disks = [
    {
      path     = var.os_disk_path
      readonly = false
      direct   = true
      id       = "rootfs"
    },
    {
      path     = var.cloudinit_disk_path
      readonly = true
      direct   = true
      id       = "cloud-init"
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

  serial = {
    file = var.serial_log_path
    mode = "File"
  }
}

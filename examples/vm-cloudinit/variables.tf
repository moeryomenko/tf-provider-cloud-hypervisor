# -----------------------------------------------------------------------------
# Variable Definitions — VM with Cloud-Init
#
# Required variables (no default):
#   - firmware_path:  Path to the firmware binary
#   - os_disk_path:   Path to the raw cloud image
#
# Optional variables with sensible defaults matching the cloud-init workflow:
#   - cloudinit_disk_path:  Path to the cloud-init FAT disk image
#   - vcpus / max_vcpus:   CPU configuration
#   - memory_size_bytes:   Memory size in bytes (NOT MiB)
#   - tap_device_name:     Host tap device for networking
#   - guest_mac_address:   Must match cloud-init/network-config
#   - serial_log_path:     Where serial console output is written
# -----------------------------------------------------------------------------

variable "firmware_path" {
  description = "Path to firmware binary (hypervisor-fw or CLOUDHV.fd)"
  type        = string
}

variable "os_disk_path" {
  description = "Path to raw cloud image (e.g., Ubuntu cloud img converted to raw)"
  type        = string
}

variable "cloudinit_disk_path" {
  description = "Path to cloud-init FAT disk image"
  type        = string
  default     = "/tmp/ubuntu-cloudinit.img"
}

variable "vcpus" {
  description = "Number of boot vCPUs (must be >= 1)"
  type        = number
  default     = 2
}

variable "max_vcpus" {
  description = "Maximum number of vCPUs (must be >= vcpus)"
  type        = number
  default     = 4
}

variable "memory_size_bytes" {
  description = "Memory size in bytes (NOT MiB). Default 2147483648 = 2 GiB."
  type        = number
  default     = 2147483648
}

variable "tap_device_name" {
  description = "Host tap device name (must exist or create with: sudo ip tuntap add ...)"
  type        = string
  default     = "chtap0"
}

variable "guest_mac_address" {
  description = "Guest MAC address (must match cloud-init/network-config macaddress)"
  type        = string
  default     = "12:34:56:78:90:ab"
}

variable "serial_log_path" {
  description = "Path to serial console log file"
  type        = string
  default     = "/tmp/ch-cloudinit-serial.log"
}

variable "kernel_path" {
  description = "Path to vmlinux kernel binary (PVH-enabled)"
  type        = string
}

variable "initrd_path" {
  description = "Path to initramfs image (optional, set to null if kernel has built-in drivers)"
  type        = string
  default     = null
}

variable "kernel_cmdline" {
  description = "Kernel command line. For direct kernel boot, specify console=, root=, and rw. The root device must match the OS disk position: root=/dev/vda1 for the first partition of the first disk."
  type        = string
  default     = "console=ttyS0 root=/dev/vda1 rw"
}

variable "os_disk_path" {
  description = "Path to OS root disk image"
  type        = string
}

variable "data_disk_path" {
  description = "Path to data disk image"
  type        = string
}

variable "vcpus" {
  description = "Number of boot vCPUs"
  type        = number
  default     = 2
}

variable "max_vcpus" {
  description = "Maximum number of vCPUs"
  type        = number
  default     = 4
}

variable "memory_size_bytes" {
  description = "Memory size in bytes (default 2147483648 = 2 GiB)"
  type        = number
  default     = 2147483648
}

variable "tap_device_name" {
  description = "Host tap device name (must be pre-created, e.g., sudo ip tuntap add name chtap0 mode tap)"
  type        = string
  default     = "chtap0"
}

variable "guest_mac_address" {
  description = "Guest MAC address"
  type        = string
  default     = "de:ad:be:ef:00:01"
}

variable "serial_log_path" {
  description = "Path to serial console log file"
  type        = string
  default     = "/tmp/ch-multidisk-serial.log"
}

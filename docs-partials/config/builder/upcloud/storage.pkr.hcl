packer {
  required_plugins {
    upcloud = {
      version = ">=v1.0.0"
      source  = "github.com/UpCloudLtd/upcloud"
    }
  }
}

variable "username" {
  type        = string
  description = "UpCloud API username"
  default     = "${env("UPCLOUD_USERNAME")}"
}

variable "password" {
  type        = string
  description = "UpCloud API password"
  default     = "${env("UPCLOUD_PASSWORD")}"
  sensitive   = true
}

source "upcloud" "test" {
  username        = "${var.username}"
  password        = "${var.password}"
  zone            = "nl-ams1"
  template_prefix = "multi-disk-server"

  # The server boots from the first storage.
  storage {
    uuid = "01000000-0000-4000-8000-000000000000"
    size = 25
  }

  # A second disk, cloned from an existing storage and grown to 50 GB. The build produces one
  # template per disk: "multi-disk-server-<timestamp>" and "multi-disk-server-<timestamp>-disk2".
  storage {
    uuid = "01000000-0000-4000-8000-000000000001"
    size = 50
  }
}

build {
  sources = ["source.upcloud.test"]

  provisioner "shell" {
    inline = [
      "lsblk",
    ]
  }
}

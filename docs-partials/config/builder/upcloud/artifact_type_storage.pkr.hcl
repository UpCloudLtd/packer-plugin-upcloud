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
  username      = "${var.username}"
  password      = "${var.password}"
  zone          = "uk-lon1"
  storage_name  = "ubuntu server 24.04"
  template_name = "ubuntu-multizone"

  # Produce detached regular storages instead of templates, so that the same artifacts can be
  # cloned into servers in any zone.
  artifact_type = "storage"
}

build {
  sources = ["source.upcloud.test"]

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get upgrade -y",
    ]
  }
}

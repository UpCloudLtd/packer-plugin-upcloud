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

variable "ssh_public_key" {
  type        = string
  description = "Path to your SSH public key file"
  default     = "~/.ssh/id_rsa.pub"
}

source "upcloud" "test" {
  username        = "${var.username}"
  password        = "${var.password}"
  zone            = "nl-ams1"
  storage_name    = "ubuntu server 20.04"
  template_prefix = "ubuntu-server"
  # uncomment to use standard tier storage
  # storage_tier = "standard
}

build {
  sources = ["source.upcloud.test"]

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get upgrade -y",
      "echo '${file(var.ssh_public_key)}' | tee /root/.ssh/authorized_keys"
    ]
  }
}

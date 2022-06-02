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
  zone            = "fi-hel1"
  storage_name    = "Debian GNU/Linux 11 (Bullseye)"
  template_prefix = "debian11"

  network_interfaces {
    ip_addresses {
      default = true
      address = "10.0.0.20"
      family  = "IPv4"
    }
    network = "<network_uuid>"
    type    = "private"
  }
  communicator = "ssh"

  # Use bastion host to get access to private network
  # ssh_bastion_username         = "<bastion_username>"
  # ssh_bastion_host             = "<bastion_host>"
  # ssh_bastion_private_key_file = "<bastion_private_key_file>"
}

build {
  sources = ["source.upcloud.test"]

  provisioner "shell" {
    environment_vars = [
      "DEBIAN_FRONTEND=noninteractive",
      "APT_LISTCHANGES_FRONTEND=none",
    ]

    inline = [
      "apt-get update",
      "apt-get upgrade -y",
      "echo '${file(var.ssh_public_key)}' | tee /root/.ssh/authorized_keys",
    ]
  }
}

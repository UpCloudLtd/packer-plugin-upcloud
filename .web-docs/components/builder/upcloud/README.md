
Type: `upcloud`

The upcloud builder is used to generate storage templates on UpCloud.

<!-- Builder Configuration Fields -->

## Required

<!-- Code generated from the comments of the Config struct in builder/upcloud/config.go; DO NOT EDIT MANUALLY -->

- `username` (string) - The username to use when interfacing with the UpCloud API.

- `password` (string) - The password to use when interfacing with the UpCloud API.

- `zone` (string) - The zone in which the server and template should be created (e.g. nl-ams1).

- `storage_uuid` (string) - The UUID of the storage you want to use as a template when creating the server.
  
  Optionally use `storage_name` parameter to find matching storage

<!-- End of code generated from the comments of the Config struct in builder/upcloud/config.go; -->


<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

## Optional

<!-- Code generated from the comments of the Config struct in builder/upcloud/config.go; DO NOT EDIT MANUALLY -->

- `storage_name` (string) - The name of the storage that will be used to find the first matching storage in the list of existing templates.
  
  Note that `storage_uuid` parameter has higher priority. You should use either `storage_uuid` or `storage_name` for not strict matching (e.g "ubuntu server 20.04").

- `template_prefix` (string) - The prefix to use for the generated template title. Defaults to `custom-image`.
  You can use this option to easily differentiate between different templates.

- `template_name` (string) - Similarly to `template_prefix`, but this will allow you to set the full template name and not just the prefix.
  Defaults to an empty string, meaning the name will be the storage title.
  You can use this option to easily differentiate between different templates.
  It cannot be used in conjunction with the prefix setting.

- `storage_size` (int) - The storage size in gigabytes. Defaults to `25`.
  Changing this value is useful if you aim to build a template for larger server configurations where the preconfigured server disk is larger than 25 GB.
  The operating system disk can also be later extended if needed. Note that Windows templates require large storage size, than default 25 Gb.

- `state_timeout_duration` (duration string | ex: "1h5m2s") - The amount of time to wait for resource state changes. Defaults to `5m`.

- `boot_wait` (duration string | ex: "1h5m2s") - The amount of time to wait after booting the server. Defaults to '0s'

- `clone_zones` ([]string) - The array of extra zones (locations) where created templates should be cloned.
  Note that default `state_timeout_duration` is not enough for cloning, better to increase a value depending on storage size.

- `network_interfaces` ([]NetworkInterface) - The array of network interfaces to request during the creation of the server for building the packer image.

- `ssh_private_key_path` (string) - Path to SSH Private Key that will be used for provisioning and stored in the template.

- `ssh_public_key_path` (string) - Path to SSH Public Key that will be used for provisioning.

<!-- End of code generated from the comments of the Config struct in builder/upcloud/config.go; -->


## Network Interfaces object (NetworkInterface)

<!-- Code generated from the comments of the NetworkInterface struct in builder/upcloud/config.go; DO NOT EDIT MANUALLY -->

- `ip_addresses` ([]IPAddress) - List of IP Addresses

- `type` (InterfaceType) - Network type (e.g. public, utility, private)

- `network` (string) - Network UUID when connecting private network

<!-- End of code generated from the comments of the NetworkInterface struct in builder/upcloud/config.go; -->


## IP Address object (IPAddress)

<!-- Code generated from the comments of the IPAddress struct in builder/upcloud/config.go; DO NOT EDIT MANUALLY -->

- `default` (bool) - Default IP address. When set to `true` SSH communicator will connect to this IP after boot.

- `family` (string) - IP address family (IPv4 or IPv6)

- `address` (string) - IP address. Note that at the moment using floating IPs is not supported.

<!-- End of code generated from the comments of the IPAddress struct in builder/upcloud/config.go; -->



<!--
  A basic example on the usage of the builder. Multiple examples
  can be provided to highlight various build configurations.

-->
## Example Usage

Here is a sample template, which you can also find in the [example](https://github.com/UpCloudLtd/packer-plugin-upcloud/tree/main/example) directory.
It reads your UpCloud API credentials from the environment variables, creates an Ubuntu 20.04 LTS server in the `nl-ams1` region and authorizes root user to loggin with your public SSH key.

```hcl
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

```

Configuration reads your SSH public key from the default location `~/.ssh/id_rsa.pub`. You can overwrite variables using command line argumen `-var`:
```sh
$ packer build -var="ssh_public_key=/some/other/path/id_rsa.pub"
```

## Network interfaces
This template uses `network_interfaces` to define network interfaces to be used during the creation of the server for building the packer image.
```hcl
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

  network_interfaces {
    ip_addresses {
      family = "IPv4"
    }
    type = "public"
  }

  network_interfaces {
    ip_addresses {
      family = "IPv4"
    }
    type = "utility"
  }
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

```

## IPv6 network interfaces
```hcl
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
      family = "IPv6"
    }
    type = "public"
  }
  communicator = "ssh"

  # Use bastion host if no IPv6 connection is available
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

```

## Private network interfaces
```hcl
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

```

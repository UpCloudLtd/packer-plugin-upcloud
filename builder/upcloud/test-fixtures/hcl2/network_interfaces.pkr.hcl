source "upcloud" "network_interfaces" {
  storage_name = "Debian GNU/Linux 11 (Bullseye)"
  storage_size = 10
  zone         = "pl-waw1"

  network_interfaces {
    ip_addresses {
      family = "IPv4"
    }
    type = "public"
  }

  network_interfaces {
    ip_addresses {
      default = true
      family  = "IPv4"
    }
    type = "utility"
  }

  communicator = "none"
  boot_wait    = "1m"
}

build {
  sources = ["source.upcloud.network_interfaces"]
}

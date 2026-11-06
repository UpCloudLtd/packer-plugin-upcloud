source "upcloud" "storage" {
  zone = "pl-waw1"

  storage {
    uuid = "01000000-0000-4000-8000-000150020100" # Rocky Linux 9
  }

  # The same public template is cloned again as a second device; the test only cares about the
  # number of disks the builder server ends up with, not about their contents.
  storage {
    uuid = "01000000-0000-4000-8000-000150020100"
    tier = "standard"
  }
}

build {
  sources = ["source.upcloud.storage"]
}


source "upcloud" "storage-uuid" {
  ssh_username    = "root"
  storage_size    = "20"
  storage_uuid    = "01000000-0000-4000-8000-000150020100" # Rocky Linux 9
  template_prefix = "test-builder"
  zone            = "pl-waw1"
}

build {
  sources = ["source.upcloud.storage-uuid"]

}

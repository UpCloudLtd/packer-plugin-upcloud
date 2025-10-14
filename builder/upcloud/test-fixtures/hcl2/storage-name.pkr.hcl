
source "upcloud" "storage-name" {
  ssh_username    = "root"
  storage_name    = "ubuntu server 24.04"
  storage_size    = "20"
  template_prefix = "test-builder"
  zone            = "pl-waw1"
}

build {
  sources = ["source.upcloud.storage-name"]

}

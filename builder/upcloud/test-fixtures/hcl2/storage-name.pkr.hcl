
source "upcloud" "storage-name" {
  ssh_username    = "root"
  storage_name    = "ubuntu server 20.04"
  storage_size    = "20"
  template_prefix = "test-builder"
  zone            = "nl-ams1"
}

build {
  sources = ["source.upcloud.storage-name"]

}


source "upcloud" "storage-uuid" {
  ssh_username    = "root"
  storage_size    = "20"
  storage_uuid    = "01000000-0000-4000-8000-000050010400"
  template_prefix = "test-builder"
  zone            = "nl-ams1"
}

build {
  sources = ["source.upcloud.storage-uuid"]

}

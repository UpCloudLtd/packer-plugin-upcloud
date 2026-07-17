
source "upcloud" "basic" {
  storage_uuid = "01000000-0000-4000-8000-000150020100" # Rocky Linux 9
  zone         = "pl-waw1"
  clone_zones  = ["de-fra1"]

  template_labels = {
    created-by  = "packer-plugin-upcloud"
    environment = "test"
  }
}

build {
  sources = ["source.upcloud.basic"]
}

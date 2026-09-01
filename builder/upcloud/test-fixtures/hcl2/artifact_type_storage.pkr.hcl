source "upcloud" "storage-artifact" {
  storage_uuid  = "01000000-0000-4000-8000-000150020100" # Rocky Linux 9
  zone          = "pl-waw1"
  artifact_type = "storage"
}

build {
  sources = ["source.upcloud.storage-artifact"]
}

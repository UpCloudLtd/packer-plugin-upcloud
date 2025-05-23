source "upcloud" "standard-tier" {
  storage_uuid = "01000000-0000-4000-8000-000150020100" # Rocky Linux 9
  zone         = "nl-ams1"
  storage_tier = "standard"
}

build {
  sources = ["source.upcloud.standard-tier"]
} 
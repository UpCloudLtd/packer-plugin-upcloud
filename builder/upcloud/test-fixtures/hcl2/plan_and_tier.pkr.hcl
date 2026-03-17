source "upcloud" "standard-tier" {
  server_plan  = "2xCPU-4GB"
  storage_uuid = "01000000-0000-4000-8000-000150020100" # Rocky Linux 9
  storage_tier = "standard"
  zone         = "pl-waw1"
}

build {
  sources = ["source.upcloud.standard-tier"]
}
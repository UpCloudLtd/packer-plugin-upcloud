# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "UpCloud"
  description = "TODO"
  identifier = "packer/UpCloudLtd/upcloud"
  component {
    type = "builder"
    name = "UpCloud"
    slug = "upcloud"
  }
  component {
    type = "post-processor"
    name = "UpCloud"
    slug = "upcloud-import"
  }
}

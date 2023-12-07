# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "UpCloud"
  description = "A builder plugin for Packer which can be used to generate storage templates on UpCloud."
  identifier = "packer/UpCloudLtd/upcloud"
  component {
    type = "builder"
    name = "UpCloud"
    slug = "upcloud"
  }
  component {
    type = "post-processor"
    name = "UpCloud Import"
    slug = "import"
  }
}

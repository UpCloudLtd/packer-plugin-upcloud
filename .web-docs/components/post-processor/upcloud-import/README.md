
Type: `upcloud-import`  
Artifact BuilderId: `packer.post-processor.upcloud-import`

The UpCloud importer can be used to import raw disk images as private templates to UpCloud.

### Required
Username and password configuration arguments can be omitted if environment variables `UPCLOUD_USERNAME` and `UPCLOUD_PASSWORD` are set.

<!-- Code generated from the comments of the Config struct in post-processor/upcloud-import/config.go; DO NOT EDIT MANUALLY -->

- `zones` ([]string) - The list of zones in which the template should be imported

- `template_name` (string) - The name of the template. Use `replace_existing` to replace existing template
  with same name or suffix template name with e.g. timestamp to avoid errors during import

<!-- End of code generated from the comments of the Config struct in post-processor/upcloud-import/config.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in post-processor/upcloud-import/config.go; DO NOT EDIT MANUALLY -->

- `username` (string) - The username to use when interfacing with the UpCloud API.

- `password` (string) - The password to use when interfacing with the UpCloud API.

- `token` (string) - The API token to use when interfacing with the UpCloud API. This is mutually exclusive with username and password.

- `replace_existing` (bool) - Replace existing template if one exists with the same name. Defaults to `false`.

- `storage_tier` (string) - The storage tier to use. Available options are `maxiops`, `archive`, and `standard`. Defaults to `maxiops`.

- `state_timeout_duration` (duration string | ex: "1h5m2s") - The amount of time to wait for resource state changes. Defaults to `60m`.

<!-- End of code generated from the comments of the Config struct in post-processor/upcloud-import/config.go; -->



### Example Usage

Import raw disk image from filesystem using `compress` post-processor to compress image before upload
```hcl
variable "username" {
  type        = string
  description = "UpCloud API username"
  default     = env("UPCLOUD_USERNAME")
}

variable "password" {
  type        = string
  description = "UpCloud API password"
  default     = env("UPCLOUD_PASSWORD")
  sensitive   = true
}

variable "image_path" {
  type        = string
  description = "Image path"
  default     = env("UPCLOUD_IMAGE_PATH")
}

source "file" "import-example" {
  source = var.image_path
  target = "tmp/${basename(var.image_path)}"
}

build {
  sources = ["file.import-example"]

  post-processors {
    post-processor "compress" {
      output = "tmp/${basename(var.image_path)}.gz"
    }
    post-processor "upcloud-import" {
      template_name    = "import-demo"
      replace_existing = true
      username         = "${var.username}"
      password         = "${var.password}"
      zones            = ["pl-waw1", "fi-hel2"]
      storage_tier     = "maxiops"
    }
  }
}

```

Import image created by QEMU builder
```hcl
source "qemu" "example" {
  format           = "raw"
  # .. rest of the parameters ..
}

build {
  sources = ["source.qemu.example"]
  post-processors {
    post-processor "upcloud-import" {
      template_name    = "${local.template_name}"
      replace_existing = true
      zones            = ["pl-waw1"]
    }
  }
}
```

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
    }
  }
}

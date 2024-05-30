variable "disk_name" {
  type     = string
}

variable "disk_size_gb" {
  type     = number
}

variable "disk_mount_path" {
  type     = string
}

variable "disk_enabled" {
    type    = bool
    default = true
}

resource "render_background_worker" "worker" {
  name    = "some-name"
  plan    = "starter"
  region  = "oregon"

  runtime_source = {
    image = {
      image_url = "nginx"
    }
  }
  disk = var.disk_enabled ? {
      name       = var.disk_name
      size_gb    = var.disk_size_gb
      mount_path = var.disk_mount_path
    } : null
}

data "render_background_worker" "worker" {
  id = render_background_worker.worker.id
}
variable "notifications_enabled" {
  type     = bool
  default = false
}

variable "preview_notifications_enabled" {
  type     = string
  default = null
}

variable "notifications_to_send" {
  type     = string
  default = null
}

resource "render_background_worker" "worker" {
  name    = "autoscaling-test"
  plan    = "starter"
  region  = "oregon"

  runtime_source = {
    image = {
      image_url = "nginx"
    }
  }

  notification_override = var.notifications_enabled ? {
      preview_notifications_enabled = var.preview_notifications_enabled
      notifications_to_send = var.notifications_to_send
    } : null
}

data "render_background_worker" "worker" {
  id = render_background_worker.worker.id
}
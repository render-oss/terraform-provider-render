variable "min" {
  type     = number
  default = null
}

variable "enabled" {
  type     = bool
  default = false
}

variable "num_instances" {
  type     = number
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

  autoscaling = var.enabled ? {
      enabled = true
      min     = var.min
      max     = 3
      criteria = {
        cpu = {
          enabled    = true
          percentage = 60
        }
        memory = {
          enabled    = false
          percentage = 70
        }
      }
    } : null

  num_instances = var.num_instances != null ? var.num_instances : null
}






data "render_background_worker" "worker" {
  id = render_background_worker.worker.id
}
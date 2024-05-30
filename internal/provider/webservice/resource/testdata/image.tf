variable "start_command" {
  type = string
  default = null
}

variable "image_url" {
  type = string
}

resource "render_web_service" "image" {
  name    = "web-service-image-tf"
  plan    = "starter"
  region  = "oregon"
  start_command = var.start_command

  runtime_source = {
    image = {
      image_url = var.image_url
    }
  }
}

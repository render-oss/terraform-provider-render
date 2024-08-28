resource "render_web_service" "web_autoscaling_test" {
  name    = "web-service-autoscaling-tf"
  plan    = "starter"
  region  = "oregon"
  runtime_source = {
    image = {
      image_url = "docker.io/library/redis"
      tag       = "latest"
    }
  }
  autoscaling = {
    enabled = true
    min     = 1
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
  }
}
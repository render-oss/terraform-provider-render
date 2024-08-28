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
}
resource "render_web_service" "web" {
  name          = "web-service-runtime-tf"
  plan          = "starter"
  region        = "oregon"

  runtime_source = {
    image = {
      image_url        = "docker.io/library/nginx:latest",
    }
  }
}

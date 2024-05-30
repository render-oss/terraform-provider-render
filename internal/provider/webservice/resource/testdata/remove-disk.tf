resource "render_web_service" "web" {
  name    = "web-service-disk-tf"
  plan    = "starter"
  region  = "oregon"
  runtime_source = {
    image = { image_url = "docker.io/library/redis:latest" }
  }
  autoscaling = {
    enabled = false
    min     = 1
    max     = 2
    criteria = {
      cpu = {
        enabled    = true
        percentage = 70
      }
      memory = {
        enabled    = true
        percentage = 70
      }
    }
  }
  env_vars = {
    "key1" = { value = "new-value" },
    "new-key" = { value = "some-value" },
  }
  secret_files = {
    "file1" = { content = "new-content" },
    "new-file" = { content = "some-content" }
  }
}
# Issue #80: enabling maintenance mode on a free-tier service must be rejected at plan
# time, since the Render API only supports it on non-free tiers.
resource "render_web_service" "service" {
  name   = "web-service-env-var-tf"
  plan   = "free"
  region = "oregon"

  runtime_source = {
    image = {
      image_url = "nginx"
    }
  }

  maintenance_mode = {
    enabled = true
  }
}

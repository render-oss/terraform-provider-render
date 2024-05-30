resource "render_web_service" "web" {
  name          = "web-service-docker-tf"
  plan          = "starter"
  region        = "oregon"

  runtime_source = {
    docker = {
      repo_url        = "https://github.com/render-examples/bun-docker",
      branch          = "main",
    }
  }
}

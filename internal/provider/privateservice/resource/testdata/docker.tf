variable "docker_command" {
  type    = string
  default = null
}

resource "render_private_service" "private" {
  name    = "some-name"
  plan    = "starter"
  region  = "oregon"
  start_command = var.docker_command
  runtime_source = {
    docker = {
      repo_url = "https://github.com/render-examples/bun-docker"
      branch   = "main"
    }
  }
}

data "render_private_service" "private" {
  id = render_private_service.private.id
}
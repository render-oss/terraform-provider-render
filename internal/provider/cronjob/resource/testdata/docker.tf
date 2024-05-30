variable "docker_command" {
  type    = string
  default = null
}

resource "render_cron_job" "cron_job" {
  name    = "some-name"
  plan    = "starter"
  region  = "oregon"
  schedule = "0 0 * * *"
  start_command = var.docker_command
  runtime_source = {
    docker = {
      repo_url = "https://github.com/render-examples/bun-docker"
      branch   = "main"
    }
  }
}

data "render_private_service" "private" {
  id = render_cron_job.cron_job.id
}
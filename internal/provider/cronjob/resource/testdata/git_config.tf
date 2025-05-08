variable "repo_url" {
  type     = string
}

variable "auto_deploy" {
  type = bool
  default = null
}

variable "auto_deploy_trigger" {
  type = string
  default = null
}

variable "paths" {
  type = string
}

variable "build_command" {
  type = string
}

variable "start_command" {
  type = string
}

resource "render_cron_job" "cron_job" {
  name    = "some-cron"
  plan    = "starter"
  region  = "oregon"
  start_command = var.start_command
  schedule = "0 0 * * *"

  runtime_source = {
    native_runtime = {
      auto_deploy   = var.auto_deploy
      auto_deploy_trigger = var.auto_deploy_trigger
      branch        = "master"
      build_command = var.build_command
      build_filter  = var.paths != null ? {
        paths         = [var.paths]
        ignored_paths = ["tests/**"]
      } : null
      repo_url      = var.repo_url
      runtime       = "node"
    }
  }
}

data "render_cron_job" "private" {
  id = render_cron_job.cron_job.id
}
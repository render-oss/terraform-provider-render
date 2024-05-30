variable "repo_url" {
  type     = string
}

variable "auto_deploy" {
  type = bool
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

resource "render_background_worker" "worker" {
  name    = "some-name"
  plan    = "starter"
  region  = "oregon"
  start_command = var.start_command

  runtime_source = {
    native_runtime = {
      auto_deploy   = var.auto_deploy
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

data "render_background_worker" "worker" {
  id = render_background_worker.worker.id
}
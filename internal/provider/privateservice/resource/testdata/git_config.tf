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

resource "render_private_service" "private" {
  name    = "some-name"
  plan    = "starter"
  region  = "oregon"
  start_command = var.start_command

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

data "render_private_service" "private" {
  id = render_private_service.private.id
}
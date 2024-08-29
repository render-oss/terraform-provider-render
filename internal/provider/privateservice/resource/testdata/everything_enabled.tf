variable "name" {
  type     = string
}

variable "plan" {
  type     = string
}

variable "region" {
  type     = string
}

variable "runtime" {
  type     = string
}

variable "previews_generation" {
  type     = string
}

variable "environment_name" {
  type     = string
  default  = null
}

variable "pre_deploy_command" {
  type     = string
  default  = null
}

locals {
  environment_map = {
    "first" = render_project.first.environments,
    "second" = render_project.second.environments,
  }
}

resource "render_project" "first"  {
  name = "first"
  environments = {
    "prod" : { name : "prod", protected_status : "protected" },
  }
}

resource "render_project" "second"  {
  name = "second"
  environments = {
    "prod" : { name : "prod", protected_status : "protected" },
  }
  # Ensure there is always an order to creating these
  depends_on = [    render_project.first  ]
}

resource "render_private_service" "private" {
  name    = var.name
  plan    = var.plan
  pre_deploy_command = var.pre_deploy_command
  region  = var.region
  runtime_source = {
    image = { image_url = "nginx" }
  }
  disk = {
    name       = "some-disk"
    size_gb    = 1
    mount_path = "/data"
  }
  previews = {
    generation = var.previews_generation
  }
  env_vars = {
    "key1" = { value = "val1" },
    "key2" = { value = "val2" },
  }
  secret_files = {
    "file1" = { content = "content1" },
    "file2" = { content = "content2" },
  }
  notification_override = {
    preview_notifications_enabled = "true"
    notifications_to_send = "all"
  }
  environment_id = var.environment_name != null ? local.environment_map[var.environment_name]["prod"].id : null
  depends_on = [render_project.first, render_project.second]
}

data "render_private_service" "private" {
  id = render_private_service.private.id
}

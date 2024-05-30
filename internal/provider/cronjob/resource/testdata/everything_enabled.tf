variable "name" {
  type = string
}

variable "plan" {
  type = string
}

variable "region" {
  type = string
}

variable "environment_name" {
  type    = string
  default = null
}

variable "schedule" {
  type = string
}

variable "env_var_value" {
  type = string
}

variable "secret_file_value" {
    type = string
}

locals {
  environment_map = {
    "first"  = render_project.first.environments,
    "second" = render_project.second.environments,
  }
}

resource "render_project" "first" {
  name         = "first"
  environments = {
    "prod" : { name : "prod", protected_status : "protected" },
  }
}

resource "render_project" "second" {
  name         = "second"
  environments = {
    "prod" : { name : "prod", protected_status : "protected" },
  }
  # Ensure there is always an order to creating these
  depends_on = [render_project.first]
}

resource "render_cron_job" "cron_job" {
  name           = var.name
  plan           = var.plan
  region         = var.region
  schedule       = var.schedule
  runtime_source = {
    image = { image_url = "nginx" }
  }
  env_vars                      = {
    "key1" = { value =  var.env_var_value },
    "key2" = { value = "val2" },
  }
  secret_files = {
    "file1" = { content = var.secret_file_value },
    "file2" = { content = "content2" },
  }
  notification_override = {
    preview_notifications_enabled = "true"
    notifications_to_send         = "all"
  }
  environment_id = var.environment_name != null ? local.environment_map[var.environment_name]["prod"].id : null
  depends_on = [render_project.first, render_project.second]
}

data "render_private_service" "cron_job" {
  id = render_cron_job.cron_job.id
}
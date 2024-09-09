variable "environment_name" {
  type = string
}
variable "has_allow_list" {
  type = bool
}
variable "max_memory_policy" {
  type = string
}
variable "name" {
  type = string
}
variable "plan" {
  type = string
}
variable "has_log_stream_setting" {
  type = string
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

resource "render_redis" "test-redis" {
  environment_id = var.environment_name != null ? local.environment_map[var.environment_name]["prod"].id : null
  max_memory_policy = var.max_memory_policy
  name = var.name
  plan = var.plan
  region = "oregon"
  ip_allow_list = var.has_allow_list ? [
    {
      cidr_block  = "1.1.1.1/32"
      description = "test"
    },
    {
      cidr_block  = "2.0.0.0/8"
      description = "test-2"
    }
  ] : null
  log_stream_override = var.has_log_stream_setting ? {
    setting = "drop"
  } : null

  depends_on = [render_project.first, render_project.second]
}
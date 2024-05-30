variable "name" {
  type = string
}

variable "database_name" {
  type = string
}

variable "database_user" {
  type = string
}

variable "high_availability_enabled" {
  type = bool
}

variable "plan" {
  type = string
}

variable "ver" {
  type = string
}

variable "read_replica" {
  type = bool
}

variable "environment_name" {
  type     = string
  default  = null
}

locals {
  environment_map = {
    "first" = render_project.first.environments
    "second" = render_project.second.environments
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

resource "render_postgres" "test" {
  name = var.name
  database_name = var.database_name
  database_user = var.database_user
  high_availability_enabled = var.high_availability_enabled
  plan = var.plan
  region = "oregon"
  version = var.ver
  read_replicas = var.read_replica ? [{
    name = "read-replica"
  }] : null

  environment_id = var.environment_name != null ? local.environment_map[var.environment_name]["prod"].id : null
  depends_on = [render_project.first, render_project.second]
}

variable "has_parameter_overrides" {
  type = bool
}

variable "has_replica" {
  type = bool
}

variable "parameter_overrides" {
  type = map(string)
  default = {}
}

variable "replica_parameter_overrides" {
  type = map(string)
  default = {}
}

resource "render_postgres" "test" {
  name    = "test-parameter-overrides"
  plan    = "pro_4gb"
  region  = "oregon"
  version = "16"

  parameter_overrides = var.has_parameter_overrides ? var.parameter_overrides : null

  read_replicas = var.has_replica ? [{
    name                = "read-replica"
    parameter_overrides = var.replica_parameter_overrides
  }] : null
}

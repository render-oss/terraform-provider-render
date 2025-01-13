variable "name2" {
  type = string
}
variable "envName" {
  type = string
}
variable "envProtStatus" {
  type = string
}

variable "networkIsolated" {
  type = bool
}

resource "render_project" "project" {
  name = var.name2
  environments = {
    prod: {
      name             = "prod"
      protected_status = "protected"
    },
    staging: {
      name             = var.envName
      protected_status = var.envProtStatus
      network_isolated = var.networkIsolated
    }
  }
}

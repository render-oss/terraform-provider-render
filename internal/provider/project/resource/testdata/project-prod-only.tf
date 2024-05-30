variable "name" {
  type = string
}

resource "render_project" "project" {
  name = var.name
  environments = {
    prod: {
        name = "prod"
        protected_status = "protected"
    }
  }
}
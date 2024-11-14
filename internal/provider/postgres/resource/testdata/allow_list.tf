variable "has_allow_list" {
  type = bool
}

resource "render_postgres" "test" {
  name = "allow-list-postgres"
  plan = "basic_256mb"
  region = "oregon"
  version = "16"
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
}

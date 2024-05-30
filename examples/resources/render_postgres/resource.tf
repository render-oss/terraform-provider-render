resource "render_postgres" "example" {
  name    = "example-postgres-instance"
  plan    = "pro"
  region  = "ohio"
  version = "14"

  database_name = "my_database"
  database_user = "my_user"

  high_availability_enabled = true
}
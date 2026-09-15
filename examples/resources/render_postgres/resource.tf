resource "render_postgres" "example" {
  name    = "example-postgres-instance"
  plan    = "1c-4g"
  region  = "ohio"
  version = "17"

  database_name = "my_database"
  database_user = "my_user"

  high_availability_enabled = true

  # Optional: Override default PostgreSQL parameters
  parameter_overrides = {
    max_connections = "200"
    shared_buffers  = "256MB"
  }

  # Optional: Configure read replicas with their own parameter overrides
  read_replicas = [{
    name = "read-replica"
    parameter_overrides = {
      max_connections = "150"
    }
  }]
}

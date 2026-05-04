resource "render_postgres" "test" {
  name = "some-name"
  database_name = "test_name_mnop"
  database_user = "test_user"
  high_availability_enabled = false
  plan = "pro_4gb"
  disk_size_gb = 20
  region = "oregon"
  version = "16"

  log_stream_override = {
    setting = "drop"
  }

  read_replicas = [{
    name = "read-replica"
    log_stream_override = {
      setting = "drop"
    }
  }]
}

data "render_postgres" "test" {
  id = render_postgres.test.id
}

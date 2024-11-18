resource "render_postgres" "test" {
  name = "some-name"
  database_name = "test_name_mnop"
  database_user = "test_user"
  high_availability_enabled = false
  plan = "basic_256mb"
  disk_size_gb = 20
  region = "oregon"
  version = "16"
}

data "render_postgres" "test" {
  id = render_postgres.test.id
}

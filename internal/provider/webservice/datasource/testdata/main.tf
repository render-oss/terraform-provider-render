resource "render_web_service" "web" {
  name    = "web-service-tf"
  plan    = "starter"
  region  = "oregon"

  runtime_source = {
    image = {
      image_url = "nginx"
    }
  }

  disk = {
    name       = "some-disk"
    size_gb    = 1
    mount_path = "/data"
  }
  env_vars = {
    "key1" = { value = "val1" },
    "key2" = { value = "val2" },
  }
  secret_files = {
    "file1" = { content = "content1" },
    "file2" = { content = "content2" },
  }
  notification_override = {
    preview_notifications_enabled = "true"
    notifications_to_send = "all"
  }
}

data "render_web_service" "web" {
  id = render_web_service.web.id
}
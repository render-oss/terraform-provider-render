resource "render_web_service" "web" {
  name                 = "new-name"
  plan                 = "standard"
  region               = "oregon"
  start_command        = "npm start"

  runtime_source = {
    native_runtime = {
      repo_url     = "https://github.com/render-examples/express-hello-world"
      branch       = "main"
      auto_deploy  = true
      build_filter = {
        paths         = ["src/**"]
        ignored_paths = ["tests/**"]
      }
      build_command = "npm install"
      runtime       = "node"
    }
  }

  disk = {
    name       = "some-disk-updated"
    size_gb    = 1
    mount_path = "/data"
  }

  env_vars = {
    "key1" = { value = "new-value" },
    "new-key" = { generate_value = true },
  }

  secret_files = {
    "file1" = { content = "new-content" },
    "new-file" = { content = "some-content" }
  }

  custom_domains = [
    { name : "terraform-provider-3.example.com" },
  ]
  notification_override = {
    preview_notifications_enabled = "true"
    notifications_to_send = "all"
  }
}

resource "render_web_service" "web" {
  name    = "web-service"
  plan    = "starter"
  region  = "ohio"
  runtime = "image"
  deploy_configuration = {
    image = {
      image_url = "docker.io/library/nginx",
      tag       = "latest",
    }
  }
}

resource "render_env_group" "envgroup" {
  name = "env-group"
  env_vars = {
    "key1" = { value = "val1" },
  }
  secret_files = {
    "file1.test" = { content = "some-content" },
  }
}

resource "render_env_group_link" "envgroup_links" {
  env_group_id = render_env_group.envgroup.id
  service_ids = [
    render_web_service.web.id,
  ]
}

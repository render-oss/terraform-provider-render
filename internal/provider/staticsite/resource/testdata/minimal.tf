locals {
  env_id = [for env in render_project.first.environments : env.id if env.name == "prod"][0]
}

resource "render_project" "first"  {
  name = "first"
  environments = {
    prod = {
      name             = "prod"
      protected_status = "protected"
    }
  }
}


resource "render_static_site" "example" {
  name          = "updated-static-site"
  repo_url      = "https://github.com/render-examples/sveltekit-static"
  build_command = "npm install && npm run build"

  branch         = "main"
  auto_deploy    = false
}
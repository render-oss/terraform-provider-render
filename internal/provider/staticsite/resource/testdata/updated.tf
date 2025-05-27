locals {
    env_id = [for env in render_project.first.environments : env.id if env.name == "prod"][0]
}

resource "render_project" "first"  {
  name = "first"
  environments = {
    prod = {
      name = "prod"
      protected_status = "protected"
    }
  }
}

variable "auto_deploy" {
  type = bool
  default = null
}

variable "auto_deploy_trigger" {
  type = string
  default = null
}

resource "render_static_site" "example" {
  name          = "updated-static-site"
  repo_url      = "https://github.com/render-examples/sveltekit-static"
  build_command = "npm install && npm run build"
  environment_id = local.env_id

  branch         = "main"
  publish_path   = "build"
  auto_deploy    = var.auto_deploy
  auto_deploy_trigger = var.auto_deploy_trigger

  build_filter = {
    paths = [
      "path/**",
    ]
    ignored_paths = [
      "node_modules/**",
    ]
  }

  env_vars = {
    SITE_TITLE = {
      value = "My Static Site"
    }
    KEY = {
      value = "value"
    }
  }

  headers = [
    {
      name  = "Cache-Control"
      value = "max-age=12345"
      path  = "/api/*"
    }
  ]

  routes = [
    {
      source      = "/about"
      destination = "/about-us"
      type        = "redirect"
    },
    {
      source      = "/blog"
      destination = "/blog/index.html"
      type        = "rewrite"
    },
  ]

  custom_domains = [
    { name : "static-site-2.example.com" },
  ]

  previews = {
    generation = "off"
  }

  notification_override = {
    preview_notifications_enabled = "true"
    notifications_to_send         = "all"
  }
}

resource "render_static_site" "example" {
  name          = "example-static-site"
  repo_url      = "https://github.com/render-examples/create-react-app"
  build_command = "npm run build"

  branch         = "master"
  publish_path   = "dist"
  root_directory = "web"
  auto_deploy    = true

  build_filter = {
    paths = [
      "src/**",
      "public/**",
    ]
    ignored_paths = [
      "tests/**",
      "docs/**",
    ]
  }

  env_vars = {
    API_URL = {
      value = "https://api.example.com"
    }
    SITE_TITLE = {
      value = "My Static Site"
    }
  }

  headers = [
    {
      name  = "X-Frame-Options"
      value = "SAMEORIGIN"
      path  = "/*"
    },
    {
      name  = "Cache-Control"
      value = "max-age=31536000"
      path  = "/assets/*"
    }
  ]

  routes = [
    {
      source      = "/blog"
      destination = "/blog/index.html"
      type        = "rewrite"
    },
    {
      source      = "/about"
      destination = "/about-us"
      type        = "redirect"
    },
    {
      source      = "/old-page"
      destination = "/new-page"
      type        = "redirect"
    },
    {
      source      = "/api/*"
      destination = "/api/index.html"
      type        = "rewrite"
    },
  ]

  custom_domains = [
    { name : "static-site.example.com" },
  ]

  pull_request_previews_enabled = true

  notification_override = {
    preview_notifications_enabled = "false"
    notifications_to_send         = "failure"
  }
}
resource "render_private_service" "example" {
  name               = "my-private-service"
  plan               = "starter"
  region             = "ohio"
  start_command      = "npm start"
  pre_deploy_command = "npm run migrate"

  runtime_source = {
    native_runtime = {
      auto_deploy   = true
      branch        = "main"
      build_command = "npm install"
      build_filter = {
        paths         = ["src/**"]
        ignored_paths = ["tests/**"]
      }
      repo_url = "https://github.com/render-examples/express-hello-world"
      runtime  = "node"
    }
  }

  autoscaling = {
    enabled = true
    min     = 1
    max     = 5
    criteria = {
      cpu = {
        enabled    = true
        percentage = 70
      }
      memory = {
        enabled    = true
        percentage = 80
      }
    }
  }

  env_vars = {
    NODE_ENV = {
      value = "production"
    }
    API_KEY = {
      value = "secret-api-key"
    }
  }

  secret_files = {
    "secrets.json" = {
      content = jsonencode({
        database_password = "secret-password"
      })
    }
  }

  num_instances = 2
  previews = {
    generation = "automatic"
  }
}

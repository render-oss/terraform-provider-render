resource "render_background_worker" "git_example" {
  name   = "git-background-worker"
  plan   = "starter"
  region = "oregon"

  start_command      = "node app.js"
  pre_deploy_command = "npm run migrate"

  runtime_source = {
    native_runtime = {
      auto_deploy   = true
      branch        = "main"
      build_command = "yarn --frozen-lockfile install"
      build_filter = {
        paths         = ["src/**"]
        ignored_paths = ["src/tests/**"]
      }
      repo_url = "https://github.com/example/git-repo"
      runtime  = "node"
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
      API_KEY = {
        value = "abc12345"
      }
      DEBUG = {
        value = "false"
      }
    }

    secret_files = {
      "secrets.json" = {
        content = jsonencode({
          username = "admin"
          password = "secret1234"
        })
      }
    }

    num_instances                 = 1
    previews = {
      generation = "off"
    }
  }
}

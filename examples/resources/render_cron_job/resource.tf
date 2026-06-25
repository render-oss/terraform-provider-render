resource "render_cron_job" "cron-job-example" {
  name          = "example-cron-job"
  plan          = "starter"
  region        = "ohio"
  schedule      = "30 2 * * *" // Run daily at 2:30 AM
  start_command = "echo 'cron job running'"

  runtime_source = {
    native_runtime = {
      auto_deploy   = true
      branch        = "main"
      build_command = "pip install -r requirements.txt"
      build_filter = {
        ignored_paths = ["cronjob/tests/**"]
        paths         = ["cronjob/**"]
      }
      repo_url = "https://github.com/render-examples/flask-hello-world"
      runtime  = "python"
    }
  }

  env_vars = {
    "key1" = { value = "val1" },
    "key2" = { value = "val2" },
  }
}
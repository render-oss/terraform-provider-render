variable "runtime" {
  type    = string
  default = null
}

resource "render_web_service" "web" {
  name          = "web-service-native-tf"
  plan          = "starter"
  region        = "oregon"
  start_command = "npm start"

  runtime_source = {
    native_runtime = {
      repo_url     = "https://github.com/render-examples/express-hello-world"
      branch       = "main"
      build_command = "npm install"
      runtime       = var.runtime
    }
  }
}

variable "env_var_count" {
  type = number
}

resource "render_background_worker" "worker" {
  name    = "env-var-test"
  plan    = "starter"
  region  = "oregon"

  runtime_source = {
    image = {
      image_url = "nginx"
    }
  }

  env_vars = (var.env_var_count == 0 ? null :
        (var.env_var_count == 1 ? {
          foo = {value = "bar"}
        } : {
          foo = {value = "bar"}
          baz = {value = "qux"}
        }))

  secret_files = (var.env_var_count == 0 ? null :
        (var.env_var_count == 1 ? {
          file1 = {content = "bar"}
        } : {
          file1 = {content = "bar"}
          file2 = {content = "qux"}
        }))
}




data "render_background_worker" "worker" {
  id = render_background_worker.worker.id
}
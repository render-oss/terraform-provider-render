variable "env_var_count" {
  type = number
}

# Issue #80: a free-tier web service. The Render API omits maintenance_mode for free
# services, so the provider must not perpetually diff on it nor send it on updates.
# The service name is kept identical to env_var.tf so the recorded by-name lookups match.
resource "render_web_service" "service" {
  name    = "web-service-env-var-tf"
  plan    = "free"
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

data "render_web_service" "service" {
  id = render_web_service.service.id
}

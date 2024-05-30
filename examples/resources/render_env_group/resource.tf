resource "render_env_group" "example" {
  name = "my-environment-group"

  env_vars = {
    DATABASE_URL = {
      value = "postgresql://user:password@localhost/mydb"
    }
    DEBUG_MODE = {
      value = "false"
    }
    INSTANCE_ID = {
      generate_value = true
    }
  }

  secret_files = {
    "secrets.json" = {
      content = jsonencode({
        aws_access_key = "AKIAIOSFODNN7EXAMPLE"
        aws_secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      })
    }
    "credentials.txt" = {
      content = "username:password"
    }
  }
}
resource "render_env_group" "import" {
  name = "some-env-group"
  env_vars = {
    "key1" = { value = "new-value" },
    "new-key" = { value = "some-value" },
  }
  secret_files = {
    "file1" = { content = "new-content" },
    "new-file" = { content = "some-content" }
  }
}
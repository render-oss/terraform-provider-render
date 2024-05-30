resource "render_registry_credential" "example" {
  name       = "my-registry-credential"
  registry   = "DOCKER"
  username   = "my-username"
  auth_token = "my-auth-token"
}
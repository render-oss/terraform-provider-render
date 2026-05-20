# Workspace-scoped: every service in the workspace within this region
# routes outbound traffic through this dedicated IP.
resource "render_dedicated_ip" "outbound" {
  name        = "primary-egress"
  description = "egress IP shared by all services in oregon"
  region      = "oregon"
}

# Environment-scoped: only services in the listed environments use this IP.
resource "render_dedicated_ip" "production" {
  name            = "production-egress"
  region          = "oregon"
  environment_ids = ["evm-abc123"]
}

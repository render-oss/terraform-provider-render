# Drops environment_ids (defaults to empty set) to switch the dedicated
# IP back to workspace-scoped. Equivalent to writing environment_ids = [].
resource "render_dedicated_ip" "example" {
  name        = "tf-acc-dsip-renamed"
  description = "updated"
  region      = "oregon"
}

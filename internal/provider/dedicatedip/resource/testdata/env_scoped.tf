# Before recording: replace the placeholder evm- ID below with a real
# environment ID from the workspace used for the recording session.
# The cassette will then capture that same ID and replay will match.
resource "render_dedicated_ip" "example" {
  name            = "tf-acc-dsip-renamed"
  description     = "updated"
  region          = "oregon"
  environment_ids = ["evm-replace-with-real-env-id"]
}

# The evm- ID is the literal value captured during cassette recording;
# don't change it without re-recording.
resource "render_dedicated_ip" "example" {
  name            = "tf-acc-dsip-renamed"
  description     = "updated"
  region          = "oregon"
  environment_ids = ["evm-d81pghf7f7vs73bocn1g"]
}

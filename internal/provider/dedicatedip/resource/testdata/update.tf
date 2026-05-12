resource "render_dedicated_ip" "example" {
  name        = "tf-acc-dsip-renamed"
  description = "updated"
  region      = "oregon"
}

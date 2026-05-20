resource "render_dedicated_ip" "src" {
  name        = "tf-acc-dsip-ds"
  description = "datasource fixture"
  region      = "oregon"
}

data "render_dedicated_ip" "read" {
  id = render_dedicated_ip.src.id
}

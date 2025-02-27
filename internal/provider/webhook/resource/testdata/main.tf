resource "render_webhook" "test" {
  name = "test-tf-webhook"
  url = "https://test-url.render.com"
  enabled = true
}

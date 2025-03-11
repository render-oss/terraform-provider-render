resource "render_metrics_stream" "settings" {
  url = "https://opentelemetry-collector-1.onrender.com"
  token = "my-token"
  metrics_provider = "CUSTOM"
}

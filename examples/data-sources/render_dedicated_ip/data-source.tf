data "render_dedicated_ip" "outbound" {
  id = "egs-abc123"
}

output "outbound_ips" {
  value = data.render_dedicated_ip.outbound.ips
}

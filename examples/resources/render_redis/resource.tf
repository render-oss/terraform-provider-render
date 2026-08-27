resource "render_redis" "example" {
  name              = "my-redis-instance"
  region            = "ohio"
  plan              = "256mb"
  max_memory_policy = "noeviction"
  persistence_mode  = "journal_snapshot"

  ip_allow_list = [
    {
      cidr_block  = "203.0.113.0/24"
      description = "Office network"
    },
  ]
}

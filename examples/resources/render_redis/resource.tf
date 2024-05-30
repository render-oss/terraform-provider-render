resource "render_redis" "example" {
  name              = "my-redis-instance"
  region            = "ohio"
  plan              = "starter"
  max_memory_policy = "noeviction"

  ip_allow_list = [
    {
      cidr_block  = "203.0.113.0/24"
      description = "Office network"
    },
  ]
}
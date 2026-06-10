variable "plan" {
  type = string
}

resource "render_redis" "test-redis-upgrade" {
  name              = "test-redis-upgrade"
  plan              = var.plan
  region            = "oregon"
  max_memory_policy = "allkeys_lfu"
}

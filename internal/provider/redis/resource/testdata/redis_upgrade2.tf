variable "plan" {
  type = string
}

variable "persistence_mode" {
  type = string
}

resource "render_redis" "test-redis-upgrade2" {
  name              = "test-redis-upgrade2"
  plan              = var.plan
  region            = "oregon"
  max_memory_policy = "allkeys_lfu"
  persistence_mode  = var.persistence_mode
}

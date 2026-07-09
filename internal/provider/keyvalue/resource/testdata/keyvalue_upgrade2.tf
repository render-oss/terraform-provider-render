variable "plan" {
  type = string
}

variable "persistence_mode" {
  type = string
}

resource "render_keyvalue" "test-keyvalue-upgrade2" {
  name              = "test-keyvalue-upgrade2"
  plan              = var.plan
  region            = "oregon"
  max_memory_policy = "allkeys_lfu"
  persistence_mode  = var.persistence_mode
}

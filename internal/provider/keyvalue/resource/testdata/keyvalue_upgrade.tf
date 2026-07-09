variable "plan" {
  type = string
}

resource "render_keyvalue" "test-keyvalue-upgrade" {
  name              = "test-keyvalue-upgrade"
  plan              = var.plan
  region            = "oregon"
  max_memory_policy = "allkeys_lfu"
}

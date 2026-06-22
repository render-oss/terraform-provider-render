resource "render_keyvalue" "test" {
  name              = "some-keyvalue"
  plan              = "starter"
  region            = "oregon"
  max_memory_policy = "noeviction"
  persistence_mode  = "snapshot"

  ip_allow_list = [
    {
      cidr_block  = "1.1.1.1/32"
      description = "one"
    },
    {
      cidr_block  = "2.2.2.2/32"
      description = "two"
    }
  ]
}

data "render_keyvalue" "test" {
  id = render_keyvalue.test.id
}

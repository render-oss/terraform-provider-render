resource "render_redis" "test" {
  name              = "some-redis"
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

data "render_redis" "test" {
  id = render_redis.test.id
}

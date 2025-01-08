---
page_title: "Migrate Redis to Key Value Resource"
subcategory: ""
description: |-
  A guide on how to migrate your terraform state from the Redis resource to a Key Value resource
---

The Key Value resource (`render_keyvalue`) is a new resource to manage your Redis and Valkey instances.
The Redis resource (`render_redis`) will not be going away, but new features will only be coming to the Key Value resource.

This guide provides steps on how to migrate your resources to the new resource definitions without losing any of your terraform state.

We'll be making use of two terraform constructs, the `removed` block to signify a resource should be removed from the terraform state,
and the `ìmport` block to import an existing resource in your Render workspace into the terraform state.

We'll take an example terraform configuration as follows with a single Redis resource

```terraform
resource "render_redis" "redistest" {
  max_memory_policy = "noeviction"
  name              = "redis-terraform"
  plan              = "starter"
  region            = "oregon"
}
```

Run a `terraform plan` to refresh your existing state, and retrieve the `ìd` of the Redis resource you are migrating.

```bash
render_redis.redistest: Refreshing state... [id=red-cud82vij1k6c73a263tg]
```

We'll want to change the resource type of this resource to a `render_keyvalue`, and then add two new blocks that inform
terraform that we would like to remove the terraform state of the original `render_redis` resource without destroying
the object in your Render workspace. We'll then import the Render resource from your workspace and place it into the terraform
state.

```terraform
resource "render_keyvalue" "redistest" {
  max_memory_policy = "noeviction"
  name              = "redis-terraform"
  plan              = "starter"
  region            = "oregon"
}

// Import the state of this Redis ID into the key value resource in terraform
import {
  to = render_keyvalue.redistest
  id = "red-cud82vij1k6c73a263tg"
}

// Signal to terraform that we would like to remove this state from terraform but
// DO NOT destroy the resource
removed {
  from = render_redis.redistest
  lifecycle {
    destroy = false
  }
}
```

Run `terraform plan` to validate the output and ensure you are not destroying any resources.

```bash
render_keyvalue.redistest: Preparing import... [id=red-cud82vij1k6c73a263tg]
render_keyvalue.redistest: Refreshing state... [id=red-cud82vij1k6c73a263tg]

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:

Terraform will perform the following actions:

  # render_keyvalue.redistest will be imported
    resource "render_keyvalue" "redistest" {
        connection_info   = (sensitive value)
        id                = "red-cud82vij1k6c73a263tg"
        ip_allow_list     = []
        max_memory_policy = "noeviction"
        name              = "redis-terraform"
        plan              = "starter"
        region            = "oregon"
    }

 # render_redis.redistest will no longer be managed by Terraform, but will not be destroyed
 # (destroy = false is set in the configuration)
 . resource "render_redis" "redistest" {
        id                = "red-cud82vij1k6c73a263tg"
        name              = "redis-terraform"
        # (5 unchanged attributes hidden)
    }

Plan: 1 to import, 0 to add, 0 to change, 0 to destroy.
╷
│ Warning: Some objects will no longer be managed by Terraform
│
│ If you apply this plan, Terraform will discard its tracking information for the following objects, but it will not delete them:
│  - render_redis.redistest
│
│ After applying this plan, Terraform will no longer manage these objects. You will need to import them into Terraform to manage them again.
```

If your output looks similar to this, with `0 destroy` in the plan output, then you can go ahead and successfully run a `terraform apply`.

You can also do this via the [`terraform state rm`](https://developer.hashicorp.com/terraform/cli/commands/state/rm) and
[`terraform import`](https://developer.hashicorp.com/terraform/cli/commands/import) commands if you would like to do this imperatively rather than declaratively.

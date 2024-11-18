# Render Provider

This is the repository of the officially supported Terraform provider for managing infrastructure on [Render](https://render.com).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Running a local version of the provider

1. Create `~/.terraformrc` with the following contents (replace `<GOPATH>` with your GOPATH)
   ```hcl
   provider_installation {
      dev_overrides {
       "registry.terraform.io/render-oss/render" = "<GOPATH>/bin"
      }

      # For all other providers, install them directly from their origin provider
      # registries as normal. If you omit this, Terraform will _only_ use
      # the dev_overrides block, and so no other providers will be available.
      direct {}
   }
   ```
2. Install the provider by running `go install`
3. Use the provider in your Terraform configuration:
   ```terraform
      terraform {
        required_providers {
          render = {
            source = "registry.terraform.io/render-oss/render"
          }
        }
      }

      provider "render" {
        api_key = "<YOUR_API_KEY>" # Alternatively, set the RENDER_API_KEY environment variable
        owner_id = "<YOUR_OWNER_ID>" # Alternatively, set the RENDER_OWNER_ID environment variable
      }
   ```

If you are successfully using a local version of the provider, you should see the following in the output of a `terraform plan`:

```shell
Warning: Provider development overrides are in effect
````

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests may create real resources, and cost money to run.

```shell
make testacc
```

### Generating cassettes for provider tests

To generate cassettes, run:

```shell
TF_ACC=1 RENDER_OWNER_ID=<your owner id such as usr-xxx or tea-xxx> RENDER_API_KEY=<your api key> RENDER_HOST=<render host such as https://api.render.com/v1 > UPDATE_RECORDINGS=true go test -count=1 -v
```

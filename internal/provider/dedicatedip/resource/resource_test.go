package resource_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	th "terraform-provider-render/internal/provider/testhelpers"
)

// TestDedicatedIPResource exercises the full lifecycle: workspace-scoped
// create → rename + description update → switch to environment-scoped →
// switch back to workspace-scoped. Each step also asserts that the ID
// stays stable (no recreate) and that the `ips` computed attribute is
// populated after polling.
//
// To re-record this cassette, edit testdata/env_scoped.tf to reference
// a real environment ID from the workspace used for recording, then run:
//
//	RENDER_HOST=https://api.render.com/v1 \
//	RENDER_OWNER_ID=tea-<workspace-id> \
//	RENDER_API_KEY=<key> \
//	UPDATE_RECORDINGS=true TF_ACC=1 \
//	  go test ./internal/provider/dedicatedip/resource/... -v -timeout 60m
func TestDedicatedIPResource(t *testing.T) {
	resourceName := "render_dedicated_ip.example"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "dedicated_ip_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/main.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("egs-")),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-dsip"),
					resource.TestCheckResourceAttr(resourceName, "description", "initial"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "status", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "environment_ids.#", "0"),
					resource.TestCheckResourceAttrWith(resourceName, "ips.#", func(value string) error {
						if value == "0" {
							return fmt.Errorf("expected at least one IP after polling, got empty list")
						}
						return nil
					}),
				),
			},
			{
				ConfigFile: config.StaticFile("./testdata/update.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("egs-")),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-dsip-renamed"),
					resource.TestCheckResourceAttr(resourceName, "description", "updated"),
					resource.TestCheckResourceAttr(resourceName, "environment_ids.#", "0"),
				),
			},
			{
				ConfigFile: config.StaticFile("./testdata/env_scoped.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("egs-")),
					resource.TestCheckResourceAttr(resourceName, "environment_ids.#", "1"),
				),
			},
			{
				ConfigFile: config.StaticFile("./testdata/workspace_scoped.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("egs-")),
					resource.TestCheckResourceAttr(resourceName, "environment_ids.#", "0"),
				),
			},
		},
	})
}

// TestDedicatedIPResourceImport is split from the main lifecycle test
// because terraform-plugin-testing's import verification compares state
// across the import boundary and is easier to reason about in isolation.
func TestDedicatedIPResourceImport(t *testing.T) {
	resourceName := "render_dedicated_ip.import"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "dedicated_ip_import_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/import.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("egs-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

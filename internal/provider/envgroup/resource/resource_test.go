package resource_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestEnvGroupResource(t *testing.T) {
	resourceName := "render_env_group.example"

	examplesPath := th.ExamplesPath(t)

	resource.Test(t, resource.TestCase{
		// testAccProtoV6ProviderFactories are used to instantiate a provider during
		// acceptance testing. The factory function will be invoked for every Terraform
		// CLI command executed to create a provider server to which the CLI can
		// reattach.
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "env_group_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile(examplesPath + "/resources/render_env_group/resource.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("evg-")),
					resource.TestCheckResourceAttr(resourceName, "name", "my-environment-group"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.DATABASE_URL.value", "postgresql://user:password@localhost/mydb"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.DEBUG_MODE.value", "false"),
					resource.TestCheckResourceAttrWith(resourceName, "env_vars.INSTANCE_ID.value", func(value string) error {
						if len(value) == 0 {
							return fmt.Errorf("value should have been generated")
						}

						return nil
					}),

					resource.TestCheckResourceAttr(resourceName, "secret_files.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.credentials.txt.content", "username:password"),
				),
			},
			{
				ConfigFile: config.StaticFile("./testdata/update.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "new-name"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "new-value"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.new-key.value", "some-value"),
					resource.TestCheckNoResourceAttr(resourceName, "env_vars.key2.value"),

					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "new-content"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.new-file.content", "some-content"),
					resource.TestCheckNoResourceAttr(resourceName, "env_vars.file2.value"),
				),
			},
		},
	})
}

// we have to test import separately because of how terraform tests compare state from imports and the fact
// that when importing an env group, we don't know if an env var was generated and just have its value
func TestEnvGroupResourceImport(t *testing.T) {
	resourceName := "render_env_group.import"

	resource.Test(t, resource.TestCase{
		// testAccProtoV6ProviderFactories are used to instantiate a provider during
		// acceptance testing. The factory function will be invoked for every Terraform
		// CLI command executed to create a provider server to which the CLI can
		// reattach.
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "env_group_import_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/import.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("evg-")),
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

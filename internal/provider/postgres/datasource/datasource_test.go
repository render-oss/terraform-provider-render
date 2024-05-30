package datasource_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestAccPostgresDataSource(t *testing.T) {
	resourceName := "data.render_postgres.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "postgres_datasource_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/postgres.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", func(value string) error {
						if !strings.HasPrefix(value, "dpg-") {
							return fmt.Errorf("expected id to start with dpg-, got: %s", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "name", "some-name"),

					resource.TestCheckResourceAttrWith(resourceName, "database_name", func(value string) error {
						if strings.HasPrefix(value, "test_name") {
							return nil
						}
						return fmt.Errorf("expected database_name to start with test_name, got: %s", value)
					}),
					resource.TestCheckResourceAttr(resourceName, "database_user", "test_user"),
					resource.TestCheckResourceAttr(resourceName, "high_availability_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "role", "primary"),
					resource.TestCheckResourceAttr(resourceName, "version", "16"),

					resource.TestCheckResourceAttrWith(resourceName, "secrets.password", func(value string) error {
						if len(value) != 32 {
							return fmt.Errorf("expected password to be 32 characters, got: %d", len(value))
						}
						return nil
					}),
				),
			},
		},
	})
}

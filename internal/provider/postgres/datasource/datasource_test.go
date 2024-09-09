package datasource_test

import (
	"fmt"
	"regexp"
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

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.password", func(value string) error {
						if len(value) != 32 {
							return fmt.Errorf("expected password to be 32 characters, got: %d", len(value))
						}
						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.internal_connection_string", func(value string) error {
						if !regexp.MustCompile(`^postgresql://test_user.*:.{32}@dpg-.*/test_name.*$`).MatchString(value) {
							return fmt.Errorf("expected internal_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.external_connection_string", func(value string) error {
						if !regexp.MustCompile(`^postgresql://test_user.*:.{32}@dpg-.*:5434/test_name.*$`).MatchString(value) {
							return fmt.Errorf("expected external_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.psql_command", func(value string) error {
						if !regexp.MustCompile(`^PGPASSWORD=.{32} psql -h dpg-.* -p 5434 -U test_user.* test_name.*$`).MatchString(value) {
							return fmt.Errorf("expected psql_command: %s to match regex", value)
						}

						return nil
					}),
				),
			},
		},
	})
}

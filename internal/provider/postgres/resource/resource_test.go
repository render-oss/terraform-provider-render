package resource_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"terraform-provider-render/internal/provider/common/checks"
	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestAccPostgresResource(t *testing.T) {
	resourceName := "render_postgres.test"

	var firstEnvironmentID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProviderConfigureWait(t, "postgres_cassette", true),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/postgres.tf"),
				ConfigVariables: config.Variables{
					"name":                      config.StringVariable("some-name"),
					"database_name":             config.StringVariable("db_name"),
					"database_user":             config.StringVariable("db_user"),
					"high_availability_enabled": config.BoolVariable(false),
					"plan":                      config.StringVariable("starter"),
					"ver":                       config.StringVariable("15"),
					"read_replica":              config.BoolVariable(false),
					"environment_name":          config.StringVariable("first"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", func(value string) error {
						if !strings.HasPrefix(value, "dpg-") {
							return fmt.Errorf("expected id to start with dpg-, got: %s", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "name", "some-name"),

					resource.TestCheckResourceAttrWith(resourceName, "database_name", func(value string) error {
						if strings.HasPrefix(value, "db_name") {
							return nil
						}
						return fmt.Errorf("expected database_name to start with db_name, got: %s", value)
					}),
					resource.TestCheckResourceAttr(resourceName, "database_user", "db_user"),

					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "version", "15"),
					resource.TestCheckResourceAttr(resourceName, "high_availability_enabled", "false"),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.password", func(value string) error {
						if len(value) != 32 {
							return fmt.Errorf("expected password to be 32 characters, got: %d", len(value))
						}
						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.internal_connection_string", func(value string) error {
						if !regexp.MustCompile("^postgres:\\/\\/db_user.*:.{32}@dpg-.*\\/db_name.*$").MatchString(value) {
							return fmt.Errorf("expected internal_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.external_connection_string", func(value string) error {
						if !regexp.MustCompile("^postgres:\\/\\/db_user.*:.{32}@dpg-.*:5434\\/db_name.*$").MatchString(value) {
							return fmt.Errorf("expected external_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.psql_command", func(value string) error {
						if !regexp.MustCompile("^PGPASSWORD=.{32} psql -h dpg-.* -p 5434 -U db_user.* db_name.*$").MatchString(value) {
							return fmt.Errorf("expected psql_command: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						firstEnvironmentID = value

						return nil
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// This can get a suffix appended
					"database_name",
				},
				ConfigVariables: config.Variables{
					"name":                      config.StringVariable("some-name"),
					"database_name":             config.StringVariable("db_name"),
					"database_user":             config.StringVariable("db_user"),
					"high_availability_enabled": config.BoolVariable(false),
					"plan":                      config.StringVariable("starter"),
					"ver":                       config.StringVariable("15"),
					"read_replica":              config.BoolVariable(false),
					"environment_name":          config.StringVariable("first"),
				},
			},
			{
				// Update fields that don't require replacement
				ConfigFile: config.StaticFile("./testdata/postgres.tf"),
				ConfigVariables: config.Variables{
					"name":                      config.StringVariable("new-name"),
					"database_name":             config.StringVariable("db_name"),
					"database_user":             config.StringVariable("db_user"),
					"high_availability_enabled": config.BoolVariable(true),
					"plan":                      config.StringVariable("pro"),
					"ver":                       config.StringVariable("15"),
					"read_replica":              config.BoolVariable(true),
					"environment_name":          config.StringVariable("second"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "new-name"),

					resource.TestCheckResourceAttr(resourceName, "plan", "pro"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "high_availability_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "read_replicas.0.name", "read-replica"),

					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						if value == firstEnvironmentID {
							return fmt.Errorf("expected a new environment_id")
						}

						return nil
					}),
				),
			},
			{
				// Update fields that require replacement
				ConfigFile: config.StaticFile("./testdata/postgres.tf"),
				ConfigVariables: config.Variables{
					"name":                      config.StringVariable("new-name"),
					"database_name":             config.StringVariable("db_name2"),
					"database_user":             config.StringVariable("db_user2"),
					"high_availability_enabled": config.BoolVariable(false),
					"plan":                      config.StringVariable("standard"),
					"ver":                       config.StringVariable("16"),
					"read_replica":              config.BoolVariable(false),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "16"),
					resource.TestCheckResourceAttrWith(resourceName, "database_name", func(value string) error {
						if strings.HasPrefix(value, "db_name2") {
							return nil
						}
						return fmt.Errorf("expected database_name to start with db_name, got: %s", value)
					}),
					resource.TestCheckResourceAttr(resourceName, "database_user", "db_user2"),
				),
			},
		},
	})
}

func TestAccPostgresIPAllowListResource(t *testing.T) {
	resourceName := "render_postgres.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "postgres_allowlist_cassette"),
		Steps: []resource.TestStep{
			{
				// create with allow list
				ConfigFile: config.StaticFile("./testdata/allow_list.tf"),
				ConfigVariables: config.Variables{
					"has_allow_list": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.cidr_block", "1.1.1.1/32"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.description", "test"),

					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.cidr_block", "2.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.description", "test-2"),
				),
			},
			{
				// remove allow list
				ConfigFile: config.StaticFile("./testdata/allow_list.tf"),
				ConfigVariables: config.Variables{
					"has_allow_list": config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.%", "0"),
				),
			},
			{
				// readd allow list
				ConfigFile: config.StaticFile("./testdata/allow_list.tf"),
				ConfigVariables: config.Variables{
					"has_allow_list": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.cidr_block", "1.1.1.1/32"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.description", "test"),

					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.cidr_block", "2.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.description", "test-2"),
				),
			},
		},
	})
}

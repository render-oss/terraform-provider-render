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

func TestKeyValueResource(t *testing.T) {
	resourceName := "render_keyvalue.test-keyvalue"

	var firstEnvironmentID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "keyvalue_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/keyvalue.tf"),
				ConfigVariables: config.Variables{
					"environment_name":       config.StringVariable("first"),
					"has_allow_list":         config.BoolVariable(true),
					"max_memory_policy":      config.StringVariable("allkeys_lfu"),
					"name":                   config.StringVariable("test-keyvalue"),
					"plan":                   config.StringVariable("starter"),
					"has_log_stream_setting": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("red-")),
					resource.TestCheckResourceAttr(resourceName, "name", "test-keyvalue"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "max_memory_policy", "allkeys_lfu"),

					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						firstEnvironmentID = value

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.cidr_block", "1.1.1.1/32"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.description", "test"),

					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.cidr_block", "2.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.description", "test-2"),

					resource.TestCheckResourceAttr(resourceName, "log_stream_override.setting", "drop"),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.internal_connection_string", func(value string) error {
						if !regexp.MustCompile(`^keyvalue://red-.*:6379$`).MatchString(value) {
							return fmt.Errorf("expected internal_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.external_connection_string", func(value string) error {
						if !regexp.MustCompile(`^keyvalues://red-[a-z0-9]+:.{32}@.*-.*.com:637[7|9]$`).MatchString(value) {
							return fmt.Errorf("expected external_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.keyvalue_cli_command", func(value string) error {
						if !regexp.MustCompile(`^ REDISCLI_AUTH=.{32} valkey-cli --user red-[a-z0-9]+ -h .*-.*.com -p 637[7|9] --tls$`).MatchString(value) {
							return fmt.Errorf("expected keyvalue_cli_command: %s to match regex", value)
						}

						return nil
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"deploy_configuration.image.image_url", // This can be expanded by the rest API
				},
				ConfigVariables: config.Variables{
					"environment_name":       config.StringVariable("first"),
					"has_allow_list":         config.BoolVariable(true),
					"max_memory_policy":      config.StringVariable("allkeys_lfu"),
					"name":                   config.StringVariable("test-keyvalue"),
					"plan":                   config.StringVariable("starter"),
					"has_log_stream_setting": config.BoolVariable(true),
				},
			},
			// Change properties that don't require a replacement
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/keyvalue.tf"),
				ConfigVariables: config.Variables{
					"environment_name":       config.StringVariable("second"),
					"has_allow_list":         config.BoolVariable(false),
					"max_memory_policy":      config.StringVariable("noeviction"),
					"name":                   config.StringVariable("test-keyvalue-new"),
					"plan":                   config.StringVariable("standard"),
					"has_log_stream_setting": config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-keyvalue-new"),
					resource.TestCheckResourceAttr(resourceName, "plan", "standard"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "max_memory_policy", "noeviction"),
					resource.TestCheckNoResourceAttr(resourceName, "log_stream_override.setting"),

					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						if value == firstEnvironmentID {
							return fmt.Errorf("expected a new environment_id")
						}

						return nil
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.%", "0"),
					),
				),
			},
			// Readd the IP allow list to ensure we can add on update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/keyvalue.tf"),
				ConfigVariables: config.Variables{
					"environment_name":       config.StringVariable("second"),
					"has_allow_list":         config.BoolVariable(true),
					"max_memory_policy":      config.StringVariable("noeviction"),
					"name":                   config.StringVariable("test-keyvalue-new"),
					"plan":                   config.StringVariable("standard"),
					"has_log_stream_setting": config.BoolVariable(false),
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

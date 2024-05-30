package resource_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"terraform-provider-render/internal/provider/common/checks"

	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestRedisResource(t *testing.T) {
	resourceName := "render_redis.test-redis"

	var firstEnvironmentID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "redis_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/redis.tf"),
				ConfigVariables: config.Variables{
					"environment_name":  config.StringVariable("first"),
					"has_allow_list":    config.BoolVariable(true),
					"max_memory_policy": config.StringVariable("allkeys_lfu"),
					"name":              config.StringVariable("test-redis"),
					"plan":              config.StringVariable("starter"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("red-")),
					resource.TestCheckResourceAttr(resourceName, "name", "test-redis"),
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
					"environment_name":  config.StringVariable("first"),
					"has_allow_list":    config.BoolVariable(true),
					"max_memory_policy": config.StringVariable("allkeys_lfu"),
					"name":              config.StringVariable("test-redis"),
					"plan":              config.StringVariable("starter"),
				},
			},
			// Change properties that don't require a replacement
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/redis.tf"),
				ConfigVariables: config.Variables{
					"environment_name":  config.StringVariable("second"),
					"has_allow_list":    config.BoolVariable(false),
					"max_memory_policy": config.StringVariable("noeviction"),
					"name":              config.StringVariable("test-redis-new"),
					"plan":              config.StringVariable("standard"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-redis-new"),
					resource.TestCheckResourceAttr(resourceName, "plan", "standard"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "max_memory_policy", "noeviction"),

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
				ConfigFile:   config.StaticFile("./testdata/redis.tf"),
				ConfigVariables: config.Variables{
					"environment_name":  config.StringVariable("second"),
					"has_allow_list":    config.BoolVariable(true),
					"max_memory_policy": config.StringVariable("noeviction"),
					"name":              config.StringVariable("test-redis-new"),
					"plan":              config.StringVariable("standard"),
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

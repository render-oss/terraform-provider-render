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

func TestAccRedisDataSource(t *testing.T) {
	resourceName := "data.render_redis.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "redis_datasource_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/redis.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", func(value string) error {
						if !strings.HasPrefix(value, "red-") {
							return fmt.Errorf("expected id to start with red-, got: %s", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "name", "some-redis"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "max_memory_policy", "noeviction"),
					resource.TestCheckResourceAttr(resourceName, "persistence_mode", "snapshot"),

					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.cidr_block", "1.1.1.1/32"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.0.description", "one"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.cidr_block", "2.2.2.2/32"),
					resource.TestCheckResourceAttr(resourceName, "ip_allow_list.1.description", "two"),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.external_connection_string", func(value string) error {
						if !regexp.MustCompile(`^rediss://red-.*:.{32}@oregon-keyvalue\..*:637[7,9]$`).MatchString(value) {
							return fmt.Errorf("expected external_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.internal_connection_string", func(value string) error {
						if !regexp.MustCompile(`^redis://red-.*:637[7,9]$`).MatchString(value) {
							return fmt.Errorf("expected internal_connection_string: %s to match regex", value)
						}

						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "connection_info.redis_cli_command", func(value string) error {
						if !regexp.MustCompile(`^ REDISCLI_AUTH=.{32} valkey-cli --user red-.* -h oregon-keyvalue\..* -p 637[7,9] --tls$`).MatchString(value) {
							return fmt.Errorf("expected cli_command: %s to match regex", value)
						}

						return nil
					}),
				),
			},
		},
	})
}

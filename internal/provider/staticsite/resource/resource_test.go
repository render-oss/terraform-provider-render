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

func TestStaticSiteResource(t *testing.T) {
	resourceName := "render_static_site.example"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "static_site_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile(th.ExamplesPath(t) + "/resources/render_static_site/resource.tf"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "example-static-site"),
					resource.TestCheckResourceAttr(resourceName, "repo_url", "https://github.com/render-examples/create-react-app"),
					resource.TestCheckResourceAttr(resourceName, "branch", "master"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy_trigger", "commit"),
					resource.TestCheckResourceAttr(resourceName, "previews.generation", "automatic"),
					resource.TestCheckResourceAttr(resourceName, "build_command", "npm run build"),
					resource.TestCheckResourceAttr(resourceName, "publish_path", "dist"),

					// build filters
					resource.TestCheckResourceAttr(resourceName, "build_filter.paths.0", "src/**"),
					resource.TestCheckResourceAttr(resourceName, "build_filter.paths.1", "public/**"),
					resource.TestCheckResourceAttr(resourceName, "build_filter.ignored_paths.0", "tests/**"),
					resource.TestCheckResourceAttr(resourceName, "build_filter.ignored_paths.1", "docs/**"),

					// env vars
					resource.TestCheckResourceAttr(resourceName, "env_vars.API_URL.value", "https://api.example.com"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.SITE_TITLE.value", "My Static Site"),

					// headers
					// Note: According to the provider docs, state value checking for sets is not necessary for
					//   non-computed attributes as the framework will automatically return test failures for configured
					//   attributes that mismatch the saved state.
					// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#TestCheckTypeSetElemAttr

					// routes
					resource.TestCheckResourceAttr(resourceName, "routes.0.source", "/blog"),
					resource.TestCheckResourceAttr(resourceName, "routes.0.destination", "/blog/index.html"),
					resource.TestCheckResourceAttr(resourceName, "routes.0.type", "rewrite"),

					resource.TestCheckResourceAttr(resourceName, "routes.1.source", "/about"),
					resource.TestCheckResourceAttr(resourceName, "routes.1.destination", "/about-us"),
					resource.TestCheckResourceAttr(resourceName, "routes.1.type", "redirect"),

					resource.TestCheckResourceAttr(resourceName, "routes.2.source", "/old-page"),
					resource.TestCheckResourceAttr(resourceName, "routes.2.destination", "/new-page"),
					resource.TestCheckResourceAttr(resourceName, "routes.2.type", "redirect"),

					resource.TestCheckResourceAttr(resourceName, "routes.3.source", "/api/*"),
					resource.TestCheckResourceAttr(resourceName, "routes.3.destination", "/api/index.html"),
					resource.TestCheckResourceAttr(resourceName, "routes.3.type", "rewrite"),

					// custom domains
					resource.TestCheckResourceAttrWith(resourceName, "custom_domains.0.id", th.CheckIDPrefix("cdm-")),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.0.name", "static-site.example.com"),

					// notifications
					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "failure"),

					resource.TestCheckResourceAttrWith(resourceName, "slug", func(value string) error {
						if !strings.HasPrefix(value, "example-static-site") {
							return fmt.Errorf("slug should start with the service name")
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/updated.tf"),
				ConfigVariables: config.Variables{
					"auto_deploy": config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-static-site"),
					resource.TestCheckResourceAttr(resourceName, "repo_url", "https://github.com/render-examples/sveltekit-static"),
					resource.TestCheckResourceAttr(resourceName, "branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy_trigger", "off"),
					resource.TestCheckResourceAttr(resourceName, "previews.generation", "off"),
					resource.TestCheckResourceAttr(resourceName, "build_command", "npm install && npm run build"),
					resource.TestCheckResourceAttr(resourceName, "publish_path", "build"),
					resource.TestCheckResourceAttrWith(resourceName, "environment_id", th.CheckIDPrefix("evm-")),

					// build filters
					resource.TestCheckResourceAttr(resourceName, "build_filter.paths.0", "path/**"),
					resource.TestCheckResourceAttr(resourceName, "build_filter.ignored_paths.0", "node_modules/**"),

					// env vars
					resource.TestCheckResourceAttr(resourceName, "env_vars.SITE_TITLE.value", "My Static Site"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.KEY.value", "value"),

					// headers
					// Note: According to the provider docs, state value checking for sets is not necessary for
					//   non-computed attributes as the framework will automatically return test failures for configured
					//   attributes that mismatch the saved state.
					// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#TestCheckTypeSetElemAttr

					// routes

					resource.TestCheckResourceAttr(resourceName, "routes.0.source", "/about"),
					resource.TestCheckResourceAttr(resourceName, "routes.0.destination", "/about-us"),
					resource.TestCheckResourceAttr(resourceName, "routes.0.type", "redirect"),

					resource.TestCheckResourceAttr(resourceName, "routes.1.source", "/blog"),
					resource.TestCheckResourceAttr(resourceName, "routes.1.destination", "/blog/index.html"),
					resource.TestCheckResourceAttr(resourceName, "routes.1.type", "rewrite"),

					// custom domains
					resource.TestCheckResourceAttrWith(resourceName, "custom_domains.0.id", th.CheckIDPrefix("cdm-")),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.0.name", "static-site-2.example.com"),

					// notifications
					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "all"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/updated.tf"),
				ConfigVariables: config.Variables{
					"auto_deploy_trigger": config.StringVariable("commit"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy_trigger", "commit"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/minimal.tf"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-static-site"),
					resource.TestCheckResourceAttr(resourceName, "repo_url", "https://github.com/render-examples/sveltekit-static"),
					resource.TestCheckResourceAttr(resourceName, "branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy_trigger", "off"),
					resource.TestCheckResourceAttr(resourceName, "previews.generation", "off"),
					resource.TestCheckResourceAttr(resourceName, "build_command", "npm install && npm run build"),
					resource.TestCheckResourceAttr(resourceName, "publish_path", "public"),
					resource.TestCheckNoResourceAttr(resourceName, "environment_id"),
					resource.TestCheckNoResourceAttr(resourceName, "build_filter"),
					resource.TestCheckNoResourceAttr(resourceName, "env_vars"),
					resource.TestCheckNoResourceAttr(resourceName, "routes"),
					resource.TestCheckNoResourceAttr(resourceName, "custom_domains"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		// create a new static site with a minimal number of fields to ensure we correctly handle default values
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "create_minimal_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/minimal.tf"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-static-site"),
					resource.TestCheckResourceAttr(resourceName, "repo_url", "https://github.com/render-examples/sveltekit-static"),
					resource.TestCheckResourceAttr(resourceName, "branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy_trigger", "off"),
					resource.TestCheckResourceAttr(resourceName, "previews.generation", "off"),
					resource.TestCheckResourceAttr(resourceName, "build_command", "npm install && npm run build"),
					resource.TestCheckResourceAttr(resourceName, "publish_path", "public"),
				),
			},
		},
	})
}

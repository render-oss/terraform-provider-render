package resource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"terraform-provider-render/internal/provider/common/checks"

	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestProject(t *testing.T) {
	resourceName := "render_project.project"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "project_cassette"),
		Steps: []resource.TestStep{
			{
				// Create project with a single environment
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/project-prod-only.tf"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("foo"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "foo"),

					resource.TestCheckResourceAttr(resourceName, "environments.prod.name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "environments.prod.protected_status", "protected"),
				),
			},
			{
				// Import project state
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ConfigVariables: config.Variables{
					"name": config.StringVariable("foo"),
				},
			},
			{
				// Change project name
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/project-prod-only.tf"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("bar"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "bar"),

					resource.TestCheckResourceAttr(resourceName, "environments.prod.name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "environments.prod.protected_status", "protected"),
				),
			},
			{
				// Add second environment to project
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/project-prod-and-staging.tf"),
				ConfigVariables: config.Variables{
					"name2":         config.StringVariable("bar"),
					"envName":       config.StringVariable("staging"),
					"envProtStatus": config.StringVariable("unprotected"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "bar"),

					resource.TestCheckResourceAttr(resourceName, "environments.prod.name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "environments.prod.protected_status", "protected"),
					resource.TestCheckResourceAttr(resourceName, "environments.staging.name", "staging"),
					resource.TestCheckResourceAttr(resourceName, "environments.staging.protected_status", "unprotected"),
				),
			},
			{
				// Change env name and protected status
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/project-prod-and-staging.tf"),
				ConfigVariables: config.Variables{
					"name2":         config.StringVariable("bar"),
					"envName":       config.StringVariable("development"),
					"envProtStatus": config.StringVariable("protected"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "bar"),

					resource.TestCheckResourceAttr(resourceName, "environments.prod.name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "environments.prod.protected_status", "protected"),
					resource.TestCheckResourceAttr(resourceName, "environments.staging.name", "development"),
					resource.TestCheckResourceAttr(resourceName, "environments.staging.protected_status", "protected"),
				),
			},
			{
				// Remove environment
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/project-prod-only.tf"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("bar"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "bar"),

					resource.TestCheckResourceAttr(resourceName, "environments.prod.name", "prod"),
					resource.TestCheckResourceAttr(resourceName, "environments.prod.protected_status", "protected"),
					resource.TestCheckNoResourceAttr(resourceName, "environments.staging"),
				),
			},
		},
	})
}

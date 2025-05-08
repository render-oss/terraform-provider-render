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

func TestCronJobResource(t *testing.T) {
	resourceName := "render_cron_job.cron_job"

	var firstEnvironmentID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "cron_job_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/everything_enabled.tf"),
				ConfigVariables: config.Variables{
					"name":                   config.StringVariable("cron-job-tf"),
					"plan":                   config.StringVariable("starter"),
					"region":                 config.StringVariable("oregon"),
					"environment_name":       config.StringVariable("first"),
					"schedule":               config.StringVariable("0 0 * * *"),
					"env_var_value":          config.StringVariable("val1"),
					"secret_file_value":      config.StringVariable("content1"),
					"has_log_stream_setting": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("crn-")),
					resource.TestCheckResourceAttr(resourceName, "name", "cron-job-tf"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "0 0 * * *"),

					resource.TestCheckResourceAttrWith(resourceName, "runtime_source.image.image_url", func(value string) error {
						hasImageName := strings.Contains(value, "nginx")
						if !hasImageName {
							return fmt.Errorf("expected image name to be present")
						}
						return nil
					}),

					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "val1"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.key2.value", "val2"),

					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "content1"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file2.content", "content2"),

					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "all"),

					resource.TestCheckResourceAttr(resourceName, "log_stream_override.setting", "drop"),

					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						firstEnvironmentID = value

						return nil
					}),
					resource.TestCheckResourceAttrWith(resourceName, "slug", func(value string) error {
						if !strings.HasPrefix(value, "cron-job-tf") {
							return fmt.Errorf("slug should start with the service name")
						}
						return nil
					}),
				),
			},
			// Change properties that don't require a replacement
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/everything_enabled.tf"),
				ConfigVariables: config.Variables{
					"name":              config.StringVariable("new-name"), // updated
					"plan":              config.StringVariable("standard"), // updated
					"region":            config.StringVariable("oregon"),
					"environment_name":  config.StringVariable("second"),    // updated
					"schedule":          config.StringVariable("0 2 * * *"), // updated
					"env_var_value":     config.StringVariable("val2"),      // updated
					"secret_file_value": config.StringVariable("content2"),  // updated
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("crn-")),
					resource.TestCheckResourceAttr(resourceName, "name", "new-name"),
					resource.TestCheckResourceAttr(resourceName, "plan", "standard"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "val2"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "content2"),
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
				),
			},
			// Remove environment
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/everything_enabled.tf"),
				ConfigVariables: config.Variables{
					"name":              config.StringVariable("new-name"),
					"plan":              config.StringVariable("standard"),
					"region":            config.StringVariable("oregon"),
					"schedule":          config.StringVariable("0 2 * * *"),
					"env_var_value":     config.StringVariable("val2"),
					"secret_file_value": config.StringVariable("content2"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "environment_id"),
				),
			},
		},
	})
}

func TestCronJobNotificationsResource(t *testing.T) {
	resourceName := "render_cron_job.cron_job"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "notifications_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/notification.tf"),
				ConfigVariables: config.Variables{
					"preview_notifications_enabled": config.StringVariable("true"),
					"notifications_to_send":         config.StringVariable("all"),
					"notifications_enabled":         config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "all"),
				),
			},
			// Change values to update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/notification.tf"),
				ConfigVariables: config.Variables{
					"preview_notifications_enabled": config.StringVariable("false"),
					"notifications_to_send":         config.StringVariable("failure"),
					"notifications_enabled":         config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "failure"),
				),
			},
			// Remove notification overrides
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/notification.tf"),
				ConfigVariables: config.Variables{
					"notifications_enabled": config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// Values should be the same as the last check. We just stopped managing them in TF
					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "failure"),
				),
			},
		},
	})
}

func TestCronJobNativeRuntimeResource(t *testing.T) {
	resourceName := "render_cron_job.cron_job"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "git_cassette"),
		Steps: []resource.TestStep{
			{
				// Create with git repo
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/git_config.tf"),
				ConfigVariables: config.Variables{
					"repo_url":      config.StringVariable("https://github.com/render-examples/express-hello-world"),
					"auto_deploy_trigger":   config.StringVariable("commit"),
					"paths":         config.StringVariable("src/**"),
					"build_command": config.StringVariable("npm install"),
					"start_command": config.StringVariable("npm start"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.repo_url", "https://github.com/render-examples/express-hello-world"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy_trigger", "commit"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_filter.paths.0", "src/**"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_command", "npm install"),
					resource.TestCheckResourceAttr(resourceName, "start_command", "npm start"),
				),
			},
			// Change values to update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/git_config.tf"),
				ConfigVariables: config.Variables{
					"repo_url":      config.StringVariable("https://github.com/render-examples/nextjs-hello-world"),
					"auto_deploy_trigger":   config.StringVariable("off"),
					"paths":         config.StringVariable("bld/**"),
					"build_command": config.StringVariable("yarn; yarn build"),
					"start_command": config.StringVariable("yarn start"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.repo_url", "https://github.com/render-examples/nextjs-hello-world"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy_trigger", "off"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_filter.paths.0", "bld/**"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_command", "yarn; yarn build"),
					resource.TestCheckResourceAttr(resourceName, "start_command", "yarn start"),
				),
			},
		},
	})
}

func TestCronJobNativeRuntimeResourceTrigger(t *testing.T) {
	resourceName := "render_cron_job.cron_job"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "git_cassette"),
		Steps: []resource.TestStep{
			{
				// Create with git repo
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/git_config.tf"),
				ConfigVariables: config.Variables{
					"repo_url":      config.StringVariable("https://github.com/render-examples/express-hello-world"),
					"auto_deploy":   config.BoolVariable(true),
					"paths":         config.StringVariable("src/**"),
					"build_command": config.StringVariable("npm install"),
					"start_command": config.StringVariable("npm start"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.repo_url", "https://github.com/render-examples/express-hello-world"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy_trigger", "commit"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_filter.paths.0", "src/**"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_command", "npm install"),
					resource.TestCheckResourceAttr(resourceName, "start_command", "npm start"),
				),
			},
			// Change values to update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/git_config.tf"),
				ConfigVariables: config.Variables{
					"repo_url":      config.StringVariable("https://github.com/render-examples/nextjs-hello-world"),
					"auto_deploy":   config.BoolVariable(false),
					"paths":         config.StringVariable("bld/**"),
					"build_command": config.StringVariable("yarn; yarn build"),
					"start_command": config.StringVariable("yarn start"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.repo_url", "https://github.com/render-examples/nextjs-hello-world"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy_trigger", "off"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_filter.paths.0", "bld/**"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_command", "yarn; yarn build"),
					resource.TestCheckResourceAttr(resourceName, "start_command", "yarn start"),
				),
			},
		},
	})
}

func TestCronJobServiceDockerResource(t *testing.T) {
	resourceName := "render_cron_job.cron_job"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "docker_cassette"),
		Steps: []resource.TestStep{
			{
				// Create with git repo
				ResourceName:    resourceName,
				ConfigFile:      config.StaticFile("./testdata/docker.tf"),
				ConfigVariables: config.Variables{},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.docker.repo_url", "https://github.com/render-examples/bun-docker"),
				),
			},
			// Change values to update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/docker.tf"),
				ConfigVariables: config.Variables{
					"docker_command": config.StringVariable("echo 'hello world'"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.docker.repo_url", "https://github.com/render-examples/bun-docker"),
					resource.TestCheckResourceAttr(resourceName, "start_command", "echo 'hello world'"),
				),
			},
		},
	})
}

func TestEnvVarsResource(t *testing.T) {
	resourceName := "render_cron_job.job"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "env_var_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/env_var.tf"),
				ConfigVariables: config.Variables{
					"env_var_count": config.IntegerVariable(0), // start with no env vars
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "env_vars"),
					resource.TestCheckNoResourceAttr(resourceName, "secret_files"),
				),
			},
			// Create env vars
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/env_var.tf"),
				ConfigVariables: config.Variables{
					"env_var_count": config.IntegerVariable(1), // add env vars
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_vars.foo.value", "bar"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "bar"),
				),
			},
			// Update env vars
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/env_var.tf"),
				ConfigVariables: config.Variables{
					"env_var_count": config.IntegerVariable(2), // update env vars
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_vars.foo.value", "bar"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "bar"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.baz.value", "qux"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file2.content", "qux"),
				),
			},
			// Remove env vars
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/env_var.tf"),
				ConfigVariables: config.Variables{
					"env_var_count": config.IntegerVariable(0), // remove env vars
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "env_vars"),
					resource.TestCheckNoResourceAttr(resourceName, "secret_files"),
				),
			},
		},
	})
}

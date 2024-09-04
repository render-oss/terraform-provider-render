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

func TestBackgroundWorkerResource(t *testing.T) {
	resourceName := "render_background_worker.worker"

	var firstEnvironmentID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "background_worker_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/everything_enabled.tf"),
				ConfigVariables: config.Variables{
					"name":                          config.StringVariable("background-worker"),
					"plan":                          config.StringVariable("starter"),
					"region":                        config.StringVariable("oregon"),
					"runtime":                       config.StringVariable("image"),
					"pull_request_previews_enabled": config.BoolVariable(true),
					"environment_name":              config.StringVariable("first"),
					"pre_deploy_command":            config.StringVariable("echo 'hello world'"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "background-worker"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),

					resource.TestCheckResourceAttrWith(resourceName, "runtime_source.image.image_url", func(value string) error {
						hasImageName := strings.Contains(value, "nginx")
						if !hasImageName {
							return fmt.Errorf("expected image name to be present")
						}
						return nil
					}),

					resource.TestCheckResourceAttrWith(resourceName, "disk.id", th.CheckIDPrefix("dsk-")),
					resource.TestCheckResourceAttr(resourceName, "disk.name", "some-disk"),
					resource.TestCheckResourceAttr(resourceName, "disk.size_gb", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk.mount_path", "/data"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "val1"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.key2.value", "val2"),

					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "content1"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file2.content", "content2"),

					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "all"),

					resource.TestCheckResourceAttr(resourceName, "pre_deploy_command", "echo 'hello world'"),

					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						firstEnvironmentID = value

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "pull_request_previews_enabled", "true"),
					resource.TestCheckResourceAttrWith(resourceName, "slug", func(value string) error {
						if !strings.HasPrefix(value, "background-worker") {
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
				ImportStateVerifyIgnore: []string{
					"runtime_source.image.image_url", // This can be expanded by the rest API
				},
				ConfigVariables: config.Variables{
					"name":                          config.StringVariable("background-worker"),
					"plan":                          config.StringVariable("starter"),
					"region":                        config.StringVariable("oregon"),
					"runtime":                       config.StringVariable("image"),
					"pull_request_previews_enabled": config.BoolVariable(true),
				},
			},
			// Change properties that don't require a replacement
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/everything_enabled.tf"),
				ConfigVariables: config.Variables{
					"name":                          config.StringVariable("new-name"), // updated
					"plan":                          config.StringVariable("standard"), // updated
					"region":                        config.StringVariable("oregon"),
					"runtime":                       config.StringVariable("image"),
					"pull_request_previews_enabled": config.BoolVariable(false),              // updated
					"environment_name":              config.StringVariable("second"),         // updated
					"pre_deploy_command":            config.StringVariable("echo 'goodbye'"), // updated
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "new-name"),
					resource.TestCheckResourceAttr(resourceName, "plan", "standard"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_previews_enabled", "false"),
					resource.TestCheckResourceAttrWith(resourceName, "environment_id", func(value string) error {
						if !strings.HasPrefix(value, "evm-") {
							return fmt.Errorf("expected environment_id to be set")
						}

						if value == firstEnvironmentID {
							return fmt.Errorf("expected a new environment_id")
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "pre_deploy_command", "echo 'goodbye'"),
				),
			},
			// Remove environment
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/everything_enabled.tf"),
				ConfigVariables: config.Variables{
					"name":                          config.StringVariable("new-name"), // updated
					"plan":                          config.StringVariable("standard"), // updated
					"region":                        config.StringVariable("oregon"),
					"runtime":                       config.StringVariable("image"),
					"pull_request_previews_enabled": config.BoolVariable(false),              // updated
					"pre_deploy_command":            config.StringVariable("echo 'goodbye'"), // updated
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

func TestBackgroundWorkerScalingResource(t *testing.T) {
	resourceName := "render_background_worker.worker"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "scaling_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/scaling.tf"),
				ConfigVariables: config.Variables{
					"min":     config.IntegerVariable(1),
					"enabled": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "autoscaling.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.min", "1"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.max", "3"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.criteria.cpu.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.criteria.cpu.percentage", "60"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.criteria.memory.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.criteria.memory.percentage", "70"),
				),
			},
			// Change min to trigger update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/scaling.tf"),
				ConfigVariables: config.Variables{
					"min":     config.IntegerVariable(2), // new value
					"enabled": config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "autoscaling.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling.min", "2"),
				),
			},
			// Turn off autoscaling
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/scaling.tf"),
				ConfigVariables: config.Variables{
					"min":     config.IntegerVariable(2), // new value
					"enabled": config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "autoscaling"),
				),
			},
			// Turn on manual scaling
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/scaling.tf"),
				ConfigVariables: config.Variables{
					"num_instances": config.IntegerVariable(2), // new value
					"enabled":       config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "num_instances", "2"),
				),
			},
		},
	})
}

func TestBackgroundWorkerNotificationsResource(t *testing.T) {
	resourceName := "render_background_worker.worker"

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

func TestBackgroundWorkerDiskResource(t *testing.T) {
	resourceName := "render_background_worker.worker"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProviderConfigureWait(t, "disks_cassette", true),
		Steps: []resource.TestStep{
			{
				// Create with disk
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/disk.tf"),
				ConfigVariables: config.Variables{
					"disk_name":       config.StringVariable("some-disk"),
					"disk_size_gb":    config.IntegerVariable(1),
					"disk_mount_path": config.StringVariable("/data"),
					"disk_enabled":    config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disk.name", "some-disk"),
					resource.TestCheckResourceAttr(resourceName, "disk.size_gb", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk.mount_path", "/data"),
				),
			},
			// Change values to update
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/disk.tf"),
				ConfigVariables: config.Variables{
					"disk_name":       config.StringVariable("new-disk-name"),
					"disk_size_gb":    config.IntegerVariable(2),
					"disk_mount_path": config.StringVariable("/data2"),
					"disk_enabled":    config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disk.name", "new-disk-name"),
					resource.TestCheckResourceAttr(resourceName, "disk.size_gb", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk.mount_path", "/data2"),
				),
			},
			// Remove disk
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/disk.tf"),
				ConfigVariables: config.Variables{
					"disk_name":       config.StringVariable("new-disk-name"),
					"disk_size_gb":    config.IntegerVariable(2),
					"disk_mount_path": config.StringVariable("/data2"),
					"disk_enabled":    config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "disk"),
				),
			},
			// Re-add disk
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/disk.tf"),
				ConfigVariables: config.Variables{
					"disk_name":       config.StringVariable("new-disk-name"),
					"disk_size_gb":    config.IntegerVariable(2),
					"disk_mount_path": config.StringVariable("/data2"),
					"disk_enabled":    config.BoolVariable(true),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disk.name", "new-disk-name"),
					resource.TestCheckResourceAttr(resourceName, "disk.size_gb", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk.mount_path", "/data2"),
				),
			},
		},
	})
}

func TestBackgroundWorkerGitConfigResource(t *testing.T) {
	resourceName := "render_background_worker.worker"

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
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_filter.paths.0", "bld/**"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.build_command", "yarn; yarn build"),
					resource.TestCheckResourceAttr(resourceName, "start_command", "yarn start"),
				),
			},
		},
	})
}

func TestBackgroundWorkerDockerResource(t *testing.T) {
	resourceName := "render_background_worker.worker"

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
	resourceName := "render_background_worker.worker"

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

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

func TestWebServiceResource(t *testing.T) {
	resourceName := "render_web_service.web"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "web_service_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile(th.ExamplesPath(t) + "/resources/render_web_service/resource.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("srv-")),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-web-service"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
					resource.TestCheckResourceAttr(resourceName, "pre_deploy_command", "echo 'hello world'"),

					// the user hasn't set num instances so we want to store it as null so they can change it outside of the provider
					resource.TestCheckNoResourceAttr(resourceName, "num_instances"),

					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.runtime", "node"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.repo_url", "https://github.com/render-examples/express-hello-world"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.auto_deploy", "true"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "val1"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.key2.value", "val2"),

					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "content1"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.file2.content", "content2"),

					resource.TestCheckResourceAttrWith(resourceName, "disk.id", th.CheckIDPrefix("dsk-")),
					resource.TestCheckResourceAttr(resourceName, "disk.name", "some-disk"),
					resource.TestCheckResourceAttr(resourceName, "disk.size_gb", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk.mount_path", "/data"),

					resource.TestCheckResourceAttrWith(resourceName, "custom_domains.0.id", th.CheckIDPrefix("cdm-")),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.0.name", "terraform-provider-1.example.com"),
					resource.TestCheckResourceAttrWith(resourceName, "custom_domains.1.id", th.CheckIDPrefix("cdm-")),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.1.name", "terraform-provider-2.example.com"),

					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "failure"),

					resource.TestCheckResourceAttrWith(resourceName, "slug", func(value string) error {
						if !strings.HasPrefix(value, "terraform-web-service") {
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
				ConfigFile:   config.StaticFile("./testdata/update.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "new-name"),

					resource.TestCheckNoResourceAttr(resourceName, "pre_deploy_command"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "new-value"),
					resource.TestCheckResourceAttrWith(resourceName, "env_vars.new-key.value", func(value string) error {
						if value == "" {
							return fmt.Errorf("value should have been generated")
						}

						return nil
					}),
					resource.TestCheckNoResourceAttr(resourceName, "env_vars.key2.value"),

					resource.TestCheckResourceAttr(resourceName, "secret_files.file1.content", "new-content"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.new-file.content", "some-content"),
					resource.TestCheckNoResourceAttr(resourceName, "env_vars.file2.value"),

					resource.TestCheckResourceAttr(resourceName, "disk.name", "some-disk-updated"),
					resource.TestCheckResourceAttr(resourceName, "disk.mount_path", "/data"),

					resource.TestCheckResourceAttr(resourceName, "custom_domains.0.name", "terraform-provider-3.example.com"),

					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "all"),
				),
			},
		},
	})
}

func TestWebServiceResource_RuntimeUpdate(t *testing.T) {
	resourceName := "render_web_service.web"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "runtime_update_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/native_runtime.tf"),
				ConfigVariables: config.Variables{
					"runtime": config.StringVariable("node"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.runtime", "node"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.repo_url", "https://github.com/render-examples/express-hello-world"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.branch", "main"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/native_runtime.tf"),
				ConfigVariables: config.Variables{
					"runtime": config.StringVariable("python"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.native_runtime.runtime", "python"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/docker_runtime.tf"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.docker.repo_url", "https://github.com/render-examples/bun-docker"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.docker.branch", "main"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/image_runtime.tf"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						checks.ExpectNoReplace(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", "docker.io/library/nginx:latest"),
				),
			},
		},
	})
}

func TestWebServiceResource_Autoscaling(t *testing.T) {
	resourceName := "render_web_service.web_autoscaling_test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "web_service_autoscaling_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/enable-autoscaling.tf"),
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
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/remove-autoscaling.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "autoscaling"),
				),
			},
		},
	})
}

func TestWebServiceResource_Image(t *testing.T) {
	resourceName := "render_web_service.image"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "web_service_image_cassette"),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/image.tf"),
				ConfigVariables: config.Variables{
					"image_url": config.StringVariable("docker.io/library/nginx:latest"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", "docker.io/library/nginx:latest"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/image.tf"),
				ConfigVariables: config.Variables{
					"image_url":     config.StringVariable("docker.io/library/nginx:stable-perl"),
					"start_command": config.StringVariable("echo hello"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", "docker.io/library/nginx:stable-perl"), // updated
					resource.TestCheckResourceAttr(resourceName, "start_command", "echo hello"),
				),
			},
			{
				ResourceName: resourceName,
				ConfigFile:   config.StaticFile("./testdata/image.tf"),
				ConfigVariables: config.Variables{
					"image_url": config.StringVariable("docker.io/library/nginx:stable-perl"),
					// removed start_command
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", "docker.io/library/nginx:stable-perl"),
					resource.TestCheckNoResourceAttr(resourceName, "start_command"),
				),
			},
		},
	})
}

func TestEnvVarsResource(t *testing.T) {
	resourceName := "render_web_service.service"

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

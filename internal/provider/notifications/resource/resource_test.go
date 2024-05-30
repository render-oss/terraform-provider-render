package resource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestNotificationResource(t *testing.T) {
	resourceName := "render_notification_setting.settings"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "notifications_resource_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/empty.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notifications_to_send", "failure"),
					resource.TestCheckResourceAttr(resourceName, "preview_notifications_enabled", "false"),
				),
			},
			{
				ConfigFile: config.StaticFile("./testdata/main.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notifications_to_send", "none"),
					resource.TestCheckResourceAttr(resourceName, "preview_notifications_enabled", "true"),
				),
			},
		},
	})
}

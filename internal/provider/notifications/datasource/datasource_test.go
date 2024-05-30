package datasource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestNotificationSettingResource(t *testing.T) {
	resourceName := "data.render_notification_setting.settings"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "notifications_datasource_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/main.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email_enabled", "true"),
				),
			},
		},
	})
}

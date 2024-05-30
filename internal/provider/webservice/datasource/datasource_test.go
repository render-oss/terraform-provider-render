package datasource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestWebServiceDataSource(t *testing.T) {
	resourceName := "data.render_web_service.web"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "webservice_datasource_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/main.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "web-service-tf"),

					resource.TestCheckResourceAttr(resourceName, "notification_override.preview_notifications_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "notification_override.notifications_to_send", "all"),
				),
			},
		},
	})
}

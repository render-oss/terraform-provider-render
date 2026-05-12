package datasource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	th "terraform-provider-render/internal/provider/testhelpers"
)

// TestDedicatedIPDataSource creates a Dedicated IP via the resource and
// then reads it back through the data source, asserting the computed
// fields surface correctly.
//
// Recording instructions match the resource test — see resource/resource_test.go.
func TestDedicatedIPDataSource(t *testing.T) {
	resourceName := "render_dedicated_ip.src"
	dataSourceName := "data.render_dedicated_ip.read"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: th.SetupRecordingProvider(t, "dedicated_ip_datasource_cassette"),
		Steps: []resource.TestStep{
			{
				ConfigFile: config.StaticFile("./testdata/main.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", th.CheckIDPrefix("egs-")),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "region", resourceName, "region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ips.#", resourceName, "ips.#"),
				),
			},
		},
	})
}

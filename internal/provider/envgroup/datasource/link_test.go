package datasource_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider"
)

func TestAccEnvGroupLinkDataSource(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		res, err := json.Marshal(&client.EnvGroup{
			Id:   "some-id",
			Name: "some-name",
			ServiceLinks: []client.ServiceLink{
				{Id: "service1"},
				{Id: "service2"},
			},
		})
		require.NoError(t, err)

		resp.Header().Set("Content-Type", "application/json")
		_, err = resp.Write(res)
		require.NoError(t, err)
	}))

	var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(fakeServer.URL))()),
	}

	resourceName := "data.render_env_group_link.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerCfg + `data "render_env_group_link" "test" { env_group_id = "some-id" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_group_id", "some-id"),

					resource.TestCheckResourceAttr(resourceName, "service_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.0", "service1"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.1", "service2"),
				),
			},
		},
	})
}

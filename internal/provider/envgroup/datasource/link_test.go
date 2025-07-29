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
			ServiceLinks: []client.EnvGroupLink{
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
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service2"),
				),
			},
		},
	})
}

func TestAccEnvGroupLinkDataSource_OrderingAndDuplicates(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		var envGroup = &client.EnvGroup{
			Id:   "order-test-id",
			Name: "order-test-name",
			// not in sorted order and with duplicates
			ServiceLinks: []client.EnvGroupLink{
				{Id: "service3"},
				{Id: "service1"},
				{Id: "service2"},
				{Id: "service1"},
			},
		}
		res, err := json.Marshal(envGroup)
		require.NoError(t, err)

		resp.Header().Set("Content-Type", "application/json")
		_, err = resp.Write(res)
		require.NoError(t, err)
	}))
	defer fakeServer.Close()

	var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(fakeServer.URL))()),
	}

	resourceName := "data.render_env_group_link.order_test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerCfg + `data "render_env_group_link" "order_test" { env_group_id = "order-test-id" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_group_id", "order-test-id"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service3"),
				),
			},
		},
	})
}

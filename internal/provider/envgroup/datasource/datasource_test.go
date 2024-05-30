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

const providerCfg = `
provider "render" {
  api_key = "some-api-key"
  owner_id = "some-owner-id"
}
`

func TestAccEnvGroupDataSource(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		res, err := json.Marshal(&client.EnvGroup{
			Id:   "some-id",
			Name: "some-name",
			EnvVars: []client.EnvVar{
				{Key: "key1", Value: "val1"},
				{Key: "key2", Value: "val2"},
			},
			SecretFiles: []client.SecretFile{
				{Name: "name1", Content: "content1"},
				{Name: "name2", Content: "content2"},
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

	resourceName := "data.render_env_group.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerCfg + `data "render_env_group" "test" { id = "some-id" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "some-id"),
					resource.TestCheckResourceAttr(resourceName, "name", "some-name"),

					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.value", "val1"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.key1.generate_value", "false"),
					resource.TestCheckResourceAttr(resourceName, "env_vars.key2.value", "val2"),

					resource.TestCheckResourceAttr(resourceName, "secret_files.name1.content", "content1"),
					resource.TestCheckResourceAttr(resourceName, "secret_files.name2.content", "content2"),
				),
			},
		},
	})
}

package resource_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/provider"

	"terraform-provider-render/internal/client"
)

const providerCfg = `
provider "render" {
  api_key = "some-api-key"
  owner_id = "some-owner-id"
}
`

func TestEnvGroupLinkResource(t *testing.T) {
	fakeServer := envGroupLinkServer(t, "some-id")

	resourceName := "render_env_group_link.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(fakeServer.URL))()),
		},
		Steps: []resource.TestStep{
			{
				Config: providerCfg + envGroupLinkConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_group_id", "some-id"),

					resource.TestCheckResourceAttr(resourceName, "service_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.0", "service1"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.1", "service2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "env_group_id",
			},
			{
				Config: providerCfg + updateEnvGroupLinkConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_group_id", "some-id"),

					resource.TestCheckResourceAttr(resourceName, "service_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.0", "service1"),
					resource.TestCheckResourceAttr(resourceName, "service_ids.1", "service-new"),
				),
			},
		},
	})
	t.Run("ExistingServiceLink", func(t *testing.T) {
		fakeServer := envGroupLinkServer(t, "existing-id")

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(fakeServer.URL))()),
			},
			Steps: []resource.TestStep{
				{
					Config: providerCfg + `
						resource "render_env_group_link" "existing" {
							env_group_id = "existing-id"
							service_ids  = ["new-service"]
						}
					`,
					ExpectError: regexp.MustCompile(`import the existing service link before adding a new service`),
				},
			},
		})
	})
}

func envGroupLinkServer(t *testing.T, envGroupID string) *httptest.Server {
	envGroup := &client.EnvGroup{Id: envGroupID}
	if envGroupID == "existing-id" {
		envGroup.ServiceLinks = []client.ServiceLink{{Id: "existing-service"}}
	}

	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json")

		switch req.Method {
		case http.MethodGet:
		case http.MethodPost:
			pathParts := strings.Split(req.URL.Path, "/")
			serviceID := pathParts[len(pathParts)-1]
			envGroup.ServiceLinks = append(envGroup.ServiceLinks, client.ServiceLink{Id: serviceID})
		case http.MethodDelete:
			pathParts := strings.Split(req.URL.Path, "/")
			serviceID := pathParts[len(pathParts)-1]

			for i, sl := range envGroup.ServiceLinks {
				if sl.Id == serviceID {
					envGroup.ServiceLinks = append(envGroup.ServiceLinks[:i], envGroup.ServiceLinks[i+1:]...)
					resp.WriteHeader(http.StatusNoContent)
					return
				}
			}
			resp.WriteHeader(http.StatusNotFound)
			return
		}

		if envGroup == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}

		res, err := json.Marshal(envGroup)
		require.NoError(t, err)

		_, err = resp.Write(res)
		require.NoError(t, err)
	}))
	return fakeServer
}

const envGroupLinkConfig = `
resource "render_env_group_link" "test" {
	env_group_id = "some-id"
	service_ids  = ["service1", "service2"]
}
`

const updateEnvGroupLinkConfig = `
resource "render_env_group_link" "test" {
	env_group_id = "some-id"
	service_ids  = ["service1", "service-new"]
}
`

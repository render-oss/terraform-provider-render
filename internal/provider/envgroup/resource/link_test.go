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
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service2"),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service-new"),
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
	t.Run("ServiceIdsOrderingAndDuplicates", func(t *testing.T) {
		fakeServer := envGroupLinkServer(t, "order-test-id")

		resourceName := "render_env_group_link.order_test"

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(fakeServer.URL))()),
			},
			Steps: []resource.TestStep{
				{
					// First apply with service1, service2 order
					Config: providerCfg + `
						resource "render_env_group_link" "order_test" {
							env_group_id = "order-test-id"
							service_ids  = ["service1", "service2", "service2"]
						}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "env_group_id", "order-test-id"),
						resource.TestCheckResourceAttr(resourceName, "service_ids.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service1"),
						resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service2"),
					),
				},
				{
					// Update with reversed order - should not cause inconsistent result error
					Config: providerCfg + `
						resource "render_env_group_link" "order_test" {
							env_group_id = "order-test-id"  
							service_ids  = ["service2", "service1"]
						}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "env_group_id", "order-test-id"),
						resource.TestCheckResourceAttr(resourceName, "service_ids.#", "2"),
						resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service1"),
						resource.TestCheckTypeSetElemAttr(resourceName, "service_ids.*", "service2"),
					),
				},
				{
					// Add a third service with mixed ordering
					Config: providerCfg + `
						resource "render_env_group_link" "order_test" {
							env_group_id = "order-test-id"
							service_ids  = ["service3", "service1", "service2", "service3"]  
						}
					`,
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
	})
}

func envGroupLinkServer(t *testing.T, envGroupID string) *httptest.Server {
	envGroup := &client.EnvGroup{Id: envGroupID}
	if envGroupID == "existing-id" {
		envGroup.ServiceLinks = []client.EnvGroupLink{{Id: "existing-service"}}
	}

	fakeServer := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json")

		switch req.Method {
		case http.MethodGet:
		case http.MethodPost:
			pathParts := strings.Split(req.URL.Path, "/")
			serviceID := pathParts[len(pathParts)-1]
			envGroup.ServiceLinks = append(envGroup.ServiceLinks, client.EnvGroupLink{Id: serviceID})
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

package resource_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider"
)

func TestAccEnvGroupLinkDeletedGroup(t *testing.T) {
	var mu sync.Mutex
	var group *client.EnvGroup
	var creates, links int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/env-groups" {
			var body client.CreateEnvGroupJSONRequestBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			creates++
			group = &client.EnvGroup{Id: fmt.Sprintf("evg-%d", creates), Name: body.Name}
			w.WriteHeader(http.StatusCreated)
		} else {
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(parts) < 2 || parts[0] != "env-groups" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if group == nil || parts[1] != group.Id {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"environment group not found"}`))
				return
			}
			switch {
			case len(parts) == 2 && r.Method == http.MethodGet:
			case len(parts) == 2 && r.Method == http.MethodDelete:
				group = nil
				w.WriteHeader(http.StatusNoContent)
				return
			case len(parts) == 4 && parts[2] == "services" && r.Method == http.MethodPost:
				links++
				group.ServiceLinks = append(group.ServiceLinks, client.EnvGroupLink{Id: parts[3]})
				w.WriteHeader(http.StatusCreated)
			case len(parts) == 4 && parts[2] == "services" && r.Method == http.MethodDelete:
				group.ServiceLinks = slices.DeleteFunc(group.ServiceLinks, func(link client.EnvGroupLink) bool { return link.Id == parts[3] })
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}
		require.NoError(t, json.NewEncoder(w).Encode(group))
	}))
	t.Cleanup(server.Close)

	const config = providerCfg + `
resource "render_env_group" "test" { name = "example" }
resource "render_env_group_link" "test" {
  env_group_id = render_env_group.test.id
  service_ids = ["srv-web", "srv-worker"]
}
`
	check := func(id string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("render_env_group.test", "id", id),
			resource.TestCheckResourceAttr("render_env_group_link.test", "env_group_id", id),
			resource.TestCheckResourceAttr("render_env_group_link.test", "service_ids.#", "2"),
			resource.TestCheckTypeSetElemAttr("render_env_group_link.test", "service_ids.*", "srv-web"),
			resource.TestCheckTypeSetElemAttr("render_env_group_link.test", "service_ids.*", "srv-worker"),
		)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(server.URL), provider.WithHTTPClient(server.Client()))()),
		},
		Steps: []resource.TestStep{
			{Config: config, Check: check("evg-1")},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					group = nil // Simulate deletion outside Terraform.
				},
				Config: config,
				Check:  check("evg-2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("render_env_group.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("render_env_group_link.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 2, creates, "the missing group should be recreated once")
	require.Equal(t, 4, links, "both services should be relinked to the new group")
	require.Nil(t, group, "test cleanup should remove the replacement group")
}

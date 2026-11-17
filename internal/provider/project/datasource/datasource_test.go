package datasource_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider"
)

func TestAccProjectDataSource(t *testing.T) {
	var mu sync.Mutex
	projectName := "example"
	environmentName := "production"
	environmentIDs := []string{"evm-prod", "evm-staging"}
	failEnvironment := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var result any
		switch r.URL.Path {
		case "/projects/prj-example":
			result = client.Project{Id: "prj-example", Name: projectName, EnvironmentIds: environmentIDs}
		case "/environments/evm-prod":
			if failEnvironment {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"environment access denied"}`))
				return
			}
			result = client.Environment{Id: "evm-prod", Name: environmentName, ProtectedStatus: client.Protected, NetworkIsolationEnabled: true}
		case "/environments/evm-staging":
			result = client.Environment{Id: "evm-staging", Name: "staging", ProtectedStatus: client.Unprotected, NetworkIsolationEnabled: false}
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		require.NoError(t, json.NewEncoder(w).Encode(result))
	}))
	t.Cleanup(server.Close)

	const config = `
provider "render" {
  api_key = "local-test-key"
  owner_id = "local-test-owner"
}
data "render_project" "test" { id = "prj-example" }
output "production_environment_id" {
  value = try(data.render_project.test.environments.production.id, "missing")
}
`
	const name = "data.render_project.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(server.URL), provider.WithHTTPClient(server.Client()))()),
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "prj-example"),
					resource.TestCheckResourceAttr(name, "name", "example"),
					resource.TestCheckResourceAttr(name, "environments.%", "2"),
					resource.TestCheckResourceAttr(name, "environments.production.id", "evm-prod"),
					resource.TestCheckResourceAttr(name, "environments.production.name", "production"),
					resource.TestCheckResourceAttr(name, "environments.production.protected_status", "protected"),
					resource.TestCheckResourceAttr(name, "environments.production.network_isolated", "true"),
					resource.TestCheckResourceAttr(name, "environments.staging.id", "evm-staging"),
					resource.TestCheckResourceAttr(name, "environments.staging.name", "staging"),
					resource.TestCheckResourceAttr(name, "environments.staging.protected_status", "unprotected"),
					resource.TestCheckResourceAttr(name, "environments.staging.network_isolated", "false"),
					resource.TestCheckOutput("production_environment_id", "evm-prod"),
				),
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					projectName = "renamed"
					environmentName = "renamed-production"
					environmentIDs = []string{"evm-prod"}
				},
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "renamed"),
					resource.TestCheckResourceAttr(name, "environments.%", "1"),
					resource.TestCheckResourceAttr(name, "environments.renamed-production.id", "evm-prod"),
					resource.TestCheckResourceAttr(name, "environments.renamed-production.name", "renamed-production"),
					resource.TestCheckNoResourceAttr(name, "environments.production"),
					resource.TestCheckNoResourceAttr(name, "environments.staging"),
					resource.TestCheckOutput("production_environment_id", "missing"),
				),
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					environmentIDs = []string{}
				},
				Config: config,
				Check:  resource.TestCheckResourceAttr(name, "environments.%", "0"),
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					environmentIDs = []string{"evm-prod"}
					failEnvironment = true
				},
				Config:      config,
				ExpectError: regexp.MustCompile("environment access denied"),
			},
		},
	})
}

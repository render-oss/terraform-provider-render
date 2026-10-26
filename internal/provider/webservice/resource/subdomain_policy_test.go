package resource_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/provider"
)

type subdomainPolicyAPI struct {
	mu           sync.Mutex
	service      map[string]any
	details      map[string]any
	writes       []map[string]any
	legacy       bool
	rejectUpdate bool
}

func (a *subdomainPolicyAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	var result any
	switch {
	case r.URL.Path == "/services" && r.Method == http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		details, ok := body["serviceDetails"].(map[string]any)
		if !ok {
			http.Error(w, "missing service details", http.StatusBadRequest)
			return
		}
		a.writes = append(a.writes, details)
		// Copy the payload so later remote changes cannot alter captured requests.
		raw, _ := json.Marshal(body)
		_ = json.Unmarshal(raw, &a.service)
		a.service["id"] = "srv-example"
		a.service["slug"] = "example"
		a.service["suspended"] = "not_suspended"
		details, _ = a.service["serviceDetails"].(map[string]any)
		a.details = details
		details["url"] = "https://example.onrender.com"
		details["healthCheckPath"] = ""
		if _, ok := details["renderSubdomainPolicy"]; !ok && !a.legacy {
			details["renderSubdomainPolicy"] = "enabled"
		}
		w.WriteHeader(http.StatusCreated)
		result = map[string]any{"service": a.service}
	case r.URL.Path == "/services" && r.Method == http.MethodGet:
		result = []any{map[string]any{"service": a.service, "cursor": "end"}}
	case r.URL.Path == "/services/srv-example" && r.Method == http.MethodGet:
		result = a.service
	case r.URL.Path == "/services/srv-example" && r.Method == http.MethodPatch:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		details, ok := body["serviceDetails"].(map[string]any)
		if !ok {
			http.Error(w, "missing service details", http.StatusBadRequest)
			return
		}
		a.writes = append(a.writes, details)
		if a.rejectUpdate {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"policy update rejected"}`))
			return
		}
		for k, v := range body {
			if k == "serviceDetails" {
				for dk, dv := range details {
					a.details[dk] = dv
				}
			} else {
				a.service[k] = v
			}
		}
		result = a.service
	case r.URL.Path == "/services/srv-example" && r.Method == http.MethodDelete:
		a.service = nil
		w.WriteHeader(http.StatusNoContent)
		return
	case r.URL.Path == "/services/srv-example/deploys" && r.Method == http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		result = map[string]any{}
	case r.URL.Path == "/services/srv-example/env-vars" || r.URL.Path == "/services/srv-example/secret-files" || r.URL.Path == "/services/srv-example/custom-domains":
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			http.Error(w, "unexpected collection method", http.StatusMethodNotAllowed)
			return
		}
		result = []any{}
	case r.URL.Path == "/notification-settings/overrides/services/srv-example":
		result = map[string]any{"notificationsToSend": "default", "previewNotificationsEnabled": "default"}
	case strings.HasSuffix(r.URL.Path, "/srv-example/log-stream"):
		w.WriteHeader(http.StatusNotFound)
		return
	default:
		http.Error(w, "unexpected request: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(result)
}

func subdomainPolicyConfig(policy, command string) string {
	return fmt.Sprintf(`
provider "render" {
  api_key = "local-test-key"
  owner_id = "local-test-owner"
}
resource "render_web_service" "test" {
  name = "example"
  plan = "starter"
  region = "oregon"
  start_command = %q
  %s
  runtime_source = {
    native_runtime = {
      repo_url = "https://github.com/example/app"
      branch = "main"
      build_command = "echo build"
      build_filter = { paths = ["src/**"], ignored_paths = ["docs/**"] }
      runtime = "node"
    }
  }
}
data "render_web_service" "test" {
  id = render_web_service.test.id
  depends_on = [render_web_service.test]
}
`, command, policy)
}

func subdomainPolicyProvider(t *testing.T, api *subdomainPolicyAPI) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	return map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(server.URL), provider.WithHTTPClient(server.Client()))()),
	}
}

func subdomainPolicyCheck(value string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("render_web_service.test", "id", "srv-example"),
		resource.TestCheckResourceAttr("render_web_service.test", "render_subdomain_policy", value),
		resource.TestCheckResourceAttr("data.render_web_service.test", "render_subdomain_policy", value),
	)
}

func TestAccWebServiceSubdomainPolicy(t *testing.T) {
	api := &subdomainPolicyAPI{}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: subdomainPolicyProvider(t, api),
		Steps: []resource.TestStep{
			{Config: subdomainPolicyConfig(`render_subdomain_policy = "disabled"`, "node server.js"), Check: subdomainPolicyCheck("disabled")},
			{ResourceName: "render_web_service.test", ImportState: true, ImportStateVerify: true},
			{Config: subdomainPolicyConfig(`render_subdomain_policy = "enabled"`, "node server.js"), Check: subdomainPolicyCheck("enabled")},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.details["renderSubdomainPolicy"] = "disabled"
				},
				Config: subdomainPolicyConfig(`render_subdomain_policy = "enabled"`, "node server.js"),
				Check:  subdomainPolicyCheck("enabled"),
			},
		},
	})
	api.mu.Lock()
	defer api.mu.Unlock()
	require.Len(t, api.writes, 3)
	for i, want := range []string{"disabled", "enabled", "enabled"} {
		require.Equal(t, want, api.writes[i]["renderSubdomainPolicy"], "request %d", i)
	}
}

func TestAccWebServiceSubdomainPolicyUnconfigured(t *testing.T) {
	for _, setting := range []string{"", "render_subdomain_policy = null"} {
		t.Run(setting, func(t *testing.T) {
			api := &subdomainPolicyAPI{}
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: subdomainPolicyProvider(t, api),
				Steps: []resource.TestStep{
					{Config: subdomainPolicyConfig(setting, "node server.js"), Check: subdomainPolicyCheck("enabled")},
					{
						PreConfig: func() {
							api.mu.Lock()
							defer api.mu.Unlock()
							api.details["renderSubdomainPolicy"] = "disabled"
						},
						Config: subdomainPolicyConfig(setting, "node updated.js"),
						Check:  subdomainPolicyCheck("disabled"),
					},
				},
			})
			api.mu.Lock()
			defer api.mu.Unlock()
			require.Len(t, api.writes, 2)
			require.NotContains(t, api.writes[0], "renderSubdomainPolicy")
			if value, ok := api.writes[1]["renderSubdomainPolicy"]; ok {
				require.Equal(t, "disabled", value, "an unrelated update must not re-enable the subdomain")
			}
		})
	}
}

func TestAccWebServiceSubdomainPolicyInvalid(t *testing.T) {
	api := &subdomainPolicyAPI{}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: subdomainPolicyProvider(t, api),
		Steps:                    []resource.TestStep{{Config: subdomainPolicyConfig(`render_subdomain_policy = "off"`, "node server.js"), ExpectError: regexp.MustCompile(`Attribute render_subdomain_policy value must be one of`)}},
	})
	require.Empty(t, api.writes, "invalid policy must be rejected before sending a request")
}

func TestAccWebServiceSubdomainPolicyLegacyResponse(t *testing.T) {
	api := &subdomainPolicyAPI{legacy: true}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: subdomainPolicyProvider(t, api),
		Steps: []resource.TestStep{{
			Config: subdomainPolicyConfig("", "node server.js"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckNoResourceAttr("render_web_service.test", "render_subdomain_policy"),
				resource.TestCheckNoResourceAttr("data.render_web_service.test", "render_subdomain_policy"),
			),
		}},
	})
}

func TestAccWebServiceSubdomainPolicyRejected(t *testing.T) {
	api := &subdomainPolicyAPI{}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: subdomainPolicyProvider(t, api),
		Steps: []resource.TestStep{
			{Config: subdomainPolicyConfig(`render_subdomain_policy = "disabled"`, "node server.js"), Check: subdomainPolicyCheck("disabled")},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.rejectUpdate = true
				},
				Config:      subdomainPolicyConfig(`render_subdomain_policy = "enabled"`, "node server.js"),
				ExpectError: regexp.MustCompile(`policy\s+update\s+rejected`),
			},
		},
	})
	api.mu.Lock()
	defer api.mu.Unlock()
	require.Len(t, api.writes, 2)
	require.Equal(t, "enabled", api.writes[1]["renderSubdomainPolicy"])
	require.Equal(t, "disabled", api.details["renderSubdomainPolicy"], "a rejected update must not change the remote policy")
}

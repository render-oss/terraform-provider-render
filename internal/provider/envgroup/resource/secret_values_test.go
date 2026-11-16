package resource_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	testresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/envgroup"
	envgroupresource "terraform-provider-render/internal/provider/envgroup/resource"
	rendertypes "terraform-provider-render/internal/provider/types"
)

func TestReadSecretValues(t *testing.T) {
	for _, tt := range []struct {
		name      string
		body      string
		wantVars  map[string]common.EnvVarModel
		wantFiles map[string]common.SecretFileModel
	}{
		{
			name: "both omitted", body: `{"id":"evg-test","name":"remote"}`,
			wantVars:  map[string]common.EnvVarModel{"TOKEN": {Value: types.StringValue("prior-token"), GenerateValue: types.BoolValue(true)}},
			wantFiles: map[string]common.SecretFileModel{"credentials": {Content: "prior-content"}},
		},
		{
			name: "env vars omitted", body: `{"id":"evg-test","name":"remote","secretFiles":[]}`,
			wantVars: map[string]common.EnvVarModel{"TOKEN": {Value: types.StringValue("prior-token"), GenerateValue: types.BoolValue(true)}},
		},
		{
			name: "secret files omitted", body: `{"id":"evg-test","name":"remote","envVars":[]}`,
			wantFiles: map[string]common.SecretFileModel{"credentials": {Content: "prior-content"}},
		},
		{
			name: "null collections", body: `{"id":"evg-test","name":"remote","envVars":null,"secretFiles":null}`,
			wantVars:  map[string]common.EnvVarModel{"TOKEN": {Value: types.StringValue("prior-token"), GenerateValue: types.BoolValue(true)}},
			wantFiles: map[string]common.SecretFileModel{"credentials": {Content: "prior-content"}},
		},
		{name: "empty collections", body: `{"id":"evg-test","name":"remote","envVars":[],"secretFiles":[]}`},
		{
			name: "populated collections", body: `{"id":"evg-test","name":"remote","envVars":[{"key":"TOKEN","value":"remote-token"}],"secretFiles":[{"name":"credentials","content":"remote-content"}]}`,
			wantVars:  map[string]common.EnvVarModel{"TOKEN": {Value: types.StringValue("remote-token"), GenerateValue: types.BoolValue(true)}},
			wantFiles: map[string]common.SecretFileModel{"credentials": {Content: "remote-content"}},
		},
		{
			name: "readable empty values", body: `{"id":"evg-test","name":"remote","envVars":[{"key":"TOKEN","value":""}],"secretFiles":[{"name":"credentials","content":""}]}`,
			wantVars:  map[string]common.EnvVarModel{"TOKEN": {Value: types.StringValue(""), GenerateValue: types.BoolValue(true)}},
			wantFiles: map[string]common.SecretFileModel{"credentials": {Content: ""}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := secretValuesResource(t, secretValuesClient(t, http.MethodGet, tt.body))
			state := secretValuesState(t)
			resp := resource.ReadResponse{State: state}

			r.Read(ctx, resource.ReadRequest{State: state}, &resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got envgroup.EnvGroupModel
			require.False(t, resp.State.Get(ctx, &got).HasError())
			assert.Equal(t, "remote", got.Name.ValueString())
			assert.Equal(t, tt.wantVars, got.EnvVars)
			assert.Equal(t, tt.wantFiles, got.SecretFiles)
		})
	}
}

func TestCreateOmittedSecretValuesUsesPlan(t *testing.T) {
	ctx := context.Background()
	r := secretValuesResource(t, secretValuesClient(t, http.MethodPost, `{"id":"evg-created","name":"created"}`))
	state := secretValuesState(t)
	plan := tfsdk.Plan(state)
	require.False(t, plan.SetAttribute(ctx, path.Root("env_vars").AtMapKey("TOKEN").AtName("generate_value"), false).HasError())
	resp := resource.CreateResponse{State: tfsdk.State{
		Schema: state.Schema,
		Raw:    tftypes.NewValue(state.Schema.Type().TerraformType(ctx), nil),
	}}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got envgroup.EnvGroupModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	assert.Equal(t, "evg-created", got.Id.ValueString())
	assert.Equal(t, "created", got.Name.ValueString())
	assert.Equal(t, map[string]common.EnvVarModel{
		"TOKEN": {Value: types.StringValue("prior-token"), GenerateValue: types.BoolValue(false)},
	}, got.EnvVars)
	assert.Equal(t, map[string]common.SecretFileModel{"credentials": {Content: "prior-content"}}, got.SecretFiles)
}

func TestReadOmittedSecretValuesWithoutPriorValues(t *testing.T) {
	ctx := context.Background()
	r := secretValuesResource(t, secretValuesClient(t, http.MethodGet, `{"id":"evg-test","name":"remote"}`))
	state := tfsdk.State{Schema: envgroupresource.EnvGroupResourceSchema(ctx)}
	require.False(t, state.Set(ctx, envgroup.EnvGroupModel{Id: types.StringValue("evg-test")}).HasError())
	resp := resource.ReadResponse{State: state}

	r.Read(ctx, resource.ReadRequest{State: state}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got envgroup.EnvGroupModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	assert.Equal(t, "remote", got.Name.ValueString())
	assert.Nil(t, got.EnvVars)
	assert.Nil(t, got.SecretFiles)
}

func TestCreateUnavailableGeneratedValueRetainsID(t *testing.T) {
	ctx := context.Background()
	r := secretValuesResource(t, secretValuesClient(t, http.MethodPost, `{"id":"evg-created","name":"created"}`))
	state := secretValuesState(t)
	plan := tfsdk.Plan(state)
	require.False(t, plan.SetAttribute(ctx, path.Root("env_vars").AtMapKey("TOKEN").AtName("value"), types.StringUnknown()).HasError())
	resp := resource.CreateResponse{State: tfsdk.State{
		Schema: state.Schema,
		Raw:    tftypes.NewValue(state.Schema.Type().TerraformType(ctx), nil),
	}}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Contains(t, resp.Diagnostics[0].Detail(), "generated")
	var got envgroup.EnvGroupModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	assert.Equal(t, "evg-created", got.Id.ValueString(), "successful creates must remain tracked when secrets are unavailable")
}

func TestUpdateOmittedSecretValuesPreservesState(t *testing.T) {
	for _, rename := range []bool{false, true} {
		method := http.MethodGet
		if rename {
			method = http.MethodPatch
		}
		t.Run(method, func(t *testing.T) {
			ctx := context.Background()
			r := secretValuesResource(t, secretValuesClient(t, method, `{"id":"evg-test","name":"remote"}`))
			state := secretValuesState(t)
			plan := tfsdk.Plan(state)
			require.False(t, plan.SetAttribute(ctx, path.Root("env_vars").AtMapKey("TOKEN").AtName("value"), types.StringUnknown()).HasError())
			if rename {
				require.False(t, plan.SetAttribute(ctx, path.Root("name"), "remote").HasError())
			}
			resp := resource.UpdateResponse{State: state}

			r.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, &resp)

			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got envgroup.EnvGroupModel
			require.False(t, resp.State.Get(ctx, &got).HasError())
			assert.Equal(t, "remote", got.Name.ValueString())
			assert.Equal(t, map[string]common.EnvVarModel{
				"TOKEN": {Value: types.StringValue("prior-token"), GenerateValue: types.BoolValue(true)},
			}, got.EnvVars)
			assert.Equal(t, map[string]common.SecretFileModel{"credentials": {Content: "prior-content"}}, got.SecretFiles)
		})
	}
}

func TestUpdateOmittedSecretValuesUsesSuccessfulPlan(t *testing.T) {
	ctx := context.Background()
	state := secretValuesState(t)
	var prior envgroup.EnvGroupModel
	require.False(t, state.Get(ctx, &prior).HasError())
	prior.EnvVars["REMOVED"] = common.EnvVarModel{Value: types.StringValue("removed-token"), GenerateValue: types.BoolValue(false)}
	prior.SecretFiles["removed-file"] = common.SecretFileModel{Content: "removed-content"}
	require.False(t, state.Set(ctx, prior).HasError())

	plan := tfsdk.Plan(state)
	require.False(t, plan.Set(ctx, envgroup.EnvGroupModel{
		Id: types.StringValue("evg-test"), Name: types.StringValue("prior"),
		EnvVars: map[string]common.EnvVarModel{
			"TOKEN": {Value: types.StringValue("updated-token"), GenerateValue: types.BoolValue(false)},
			"ADDED": {Value: types.StringValue("added-token"), GenerateValue: types.BoolValue(false)},
		},
		SecretFiles: map[string]common.SecretFileModel{
			"credentials": {Content: "updated-content"},
			"added-file":  {Content: "added-content"},
		},
	}).HasError())

	wantRequests := map[string]string{
		"PUT /env-groups/evg-test/env-vars/TOKEN":               `{"value":"updated-token"}`,
		"PUT /env-groups/evg-test/env-vars/ADDED":               `{"value":"added-token"}`,
		"DELETE /env-groups/evg-test/env-vars/REMOVED":          "",
		"PUT /env-groups/evg-test/secret-files/credentials":     `{"content":"updated-content"}`,
		"PUT /env-groups/evg-test/secret-files/added-file":      `{"content":"added-content"}`,
		"DELETE /env-groups/evg-test/secret-files/removed-file": "",
		"GET /env-groups/evg-test":                              "",
	}
	var mu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.Method + " " + r.URL.Path
		mu.Lock()
		requests = append(requests, request)
		mu.Unlock()
		wantBody, ok := wantRequests[request]
		assert.True(t, ok, "unexpected request: %s", request)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if wantBody != "" {
			assert.JSONEq(t, wantBody, string(body))
		} else {
			assert.Empty(t, body)
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(w, `{"id":"evg-test","name":"prior"}`)
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	apiClient, err := client.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	r := secretValuesResource(t, apiClient)
	resp := resource.UpdateResponse{State: state}

	r.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got envgroup.EnvGroupModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	assert.Equal(t, map[string]common.EnvVarModel{
		"TOKEN": {Value: types.StringValue("updated-token"), GenerateValue: types.BoolValue(false)},
		"ADDED": {Value: types.StringValue("added-token"), GenerateValue: types.BoolValue(false)},
	}, got.EnvVars)
	assert.Equal(t, map[string]common.SecretFileModel{
		"credentials": {Content: "updated-content"},
		"added-file":  {Content: "added-content"},
	}, got.SecretFiles)
	mu.Lock()
	defer mu.Unlock()
	assert.ElementsMatch(t, []string{
		"PUT /env-groups/evg-test/env-vars/TOKEN",
		"PUT /env-groups/evg-test/env-vars/ADDED",
		"DELETE /env-groups/evg-test/env-vars/REMOVED",
		"PUT /env-groups/evg-test/secret-files/credentials",
		"PUT /env-groups/evg-test/secret-files/added-file",
		"DELETE /env-groups/evg-test/secret-files/removed-file",
		"GET /env-groups/evg-test",
	}, requests)
}

func TestEnvGroupResourceDeleteOmittedSecrets(t *testing.T) {
	const resourceName = "render_env_group.test"
	const populatedResponse = `{
		"id":"evg-test","name":"test",
		"envVars":[{"key":"FIRST","value":"first-value"},{"key":"SECOND","value":"second-value"}],
		"secretFiles":[{"name":"first-file","content":"first-content"},{"name":"second-file","content":"second-content"}]
	}`
	const initialConfig = providerCfg + `
resource "render_env_group" "test" {
  name = "test"
  env_vars = {
    FIRST  = { value = "first-value" }
    SECOND = { value = "second-value" }
  }
  secret_files = {
    first-file  = { content = "first-content" }
    second-file = { content = "second-content" }
  }
}
`
	const partialConfig = providerCfg + `
resource "render_env_group" "test" {
  name = "test"
  env_vars = {
    SECOND = { value = "second-value" }
  }
  secret_files = {
    second-file = { content = "second-content" }
  }
}
`
	const removedConfig = providerCfg + `
resource "render_env_group" "test" {
  name = "test"
}
`

	var mu sync.Mutex
	var omitSecrets bool
	var omittedReads int
	var deletedPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		body := populatedResponse
		switch r.Method + " " + r.URL.Path {
		case "POST /env-groups":
			w.WriteHeader(http.StatusCreated)
		case "GET /env-groups/evg-test":
			if omitSecrets {
				omittedReads++
				body = `{"id":"evg-test","name":"test"}`
			}
		case "DELETE /env-groups/evg-test/env-vars/FIRST",
			"DELETE /env-groups/evg-test/env-vars/SECOND",
			"DELETE /env-groups/evg-test/secret-files/first-file",
			"DELETE /env-groups/evg-test/secret-files/second-file":
			deletedPaths = append(deletedPaths, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		case "DELETE /env-groups/evg-test":
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		_, err := io.WriteString(w, body)
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	initialStateCheck := testresource.ComposeAggregateTestCheckFunc(
		testresource.TestCheckResourceAttr(resourceName, "env_vars.%", "2"),
		testresource.TestCheckResourceAttr(resourceName, "env_vars.FIRST.value", "first-value"),
		testresource.TestCheckResourceAttr(resourceName, "env_vars.SECOND.value", "second-value"),
		testresource.TestCheckResourceAttr(resourceName, "secret_files.%", "2"),
		testresource.TestCheckResourceAttr(resourceName, "secret_files.first-file.content", "first-content"),
		testresource.TestCheckResourceAttr(resourceName, "secret_files.second-file.content", "second-content"),
	)
	deletionPlanChecks := testresource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
		PostApplyPreRefresh:  []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}

	testresource.Test(t, testresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"render": providerserver.NewProtocol6WithError(provider.New("test", provider.WithHost(server.URL))()),
		},
		Steps: []testresource.TestStep{
			{
				Config: initialConfig,
				Check:  initialStateCheck,
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					omitSecrets = true
				},
				RefreshState: true,
				Check:        initialStateCheck,
				RefreshPlanChecks: testresource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					assert.Positive(t, omittedReads, "refresh must exercise an omitted-secrets response")
					assert.Empty(t, deletedPaths)
				},
				Config:           partialConfig,
				ConfigPlanChecks: deletionPlanChecks,
				Check: testresource.ComposeAggregateTestCheckFunc(
					testresource.TestCheckResourceAttr(resourceName, "env_vars.%", "1"),
					testresource.TestCheckNoResourceAttr(resourceName, "env_vars.FIRST.value"),
					testresource.TestCheckResourceAttr(resourceName, "env_vars.SECOND.value", "second-value"),
					testresource.TestCheckResourceAttr(resourceName, "secret_files.%", "1"),
					testresource.TestCheckNoResourceAttr(resourceName, "secret_files.first-file.content"),
					testresource.TestCheckResourceAttr(resourceName, "secret_files.second-file.content", "second-content"),
				),
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					assert.ElementsMatch(t, []string{
						"/env-groups/evg-test/env-vars/FIRST",
						"/env-groups/evg-test/secret-files/first-file",
					}, deletedPaths)
				},
				Config:           removedConfig,
				ConfigPlanChecks: deletionPlanChecks,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("env_vars"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("secret_files"), knownvalue.Null()),
				},
			},
		},
	})

	mu.Lock()
	defer mu.Unlock()
	assert.ElementsMatch(t, []string{
		"/env-groups/evg-test/env-vars/FIRST",
		"/env-groups/evg-test/env-vars/SECOND",
		"/env-groups/evg-test/secret-files/first-file",
		"/env-groups/evg-test/secret-files/second-file",
	}, deletedPaths)
}

func secretValuesResource(t *testing.T, apiClient *client.ClientWithResponses) resource.Resource {
	t.Helper()
	r, ok := envgroupresource.NewEnvGroupResource().(resource.ResourceWithConfigure)
	require.True(t, ok)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: &rendertypes.Data{Client: apiClient, OwnerID: "tea-test"},
	}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return r
}

func secretValuesState(t *testing.T) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	state := tfsdk.State{Schema: envgroupresource.EnvGroupResourceSchema(ctx)}
	diags := state.Set(ctx, envgroup.EnvGroupModel{
		Id:   types.StringValue("evg-test"),
		Name: types.StringValue("prior"),
		EnvVars: map[string]common.EnvVarModel{
			"TOKEN": {Value: types.StringValue("prior-token"), GenerateValue: types.BoolValue(true)},
		},
		SecretFiles: map[string]common.SecretFileModel{
			"credentials": {Content: "prior-content"},
		},
	})
	require.False(t, diags.HasError(), "%v", diags)
	return state
}

func secretValuesClient(t *testing.T, method, body string) *client.ClientWithResponses {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, method, r.Method)
		w.Header().Set("Content-Type", "application/json")
		if method == http.MethodPost {
			assert.Equal(t, "/env-groups", r.URL.Path)
			w.WriteHeader(http.StatusCreated)
		} else {
			assert.Equal(t, "/env-groups/evg-test", r.URL.Path)
		}
		_, err := io.WriteString(w, body)
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	apiClient, err := client.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	return apiClient
}

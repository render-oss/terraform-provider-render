package resource

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/envgroup"
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
			r := &envGroupResource{client: secretValuesClient(t, http.MethodGet, tt.body)}
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
	r := &envGroupResource{client: secretValuesClient(t, http.MethodPost, `{"id":"evg-created","name":"created"}`), ownerID: "tea-test"}
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
	r := &envGroupResource{client: secretValuesClient(t, http.MethodGet, `{"id":"evg-test","name":"remote"}`)}
	state := tfsdk.State{Schema: EnvGroupResourceSchema(ctx)}
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
	r := &envGroupResource{client: secretValuesClient(t, http.MethodPost, `{"id":"evg-created","name":"created"}`), ownerID: "tea-test"}
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
			r := &envGroupResource{client: secretValuesClient(t, method, `{"id":"evg-test","name":"remote"}`)}
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
	r := &envGroupResource{client: apiClient}
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

func secretValuesState(t *testing.T) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	state := tfsdk.State{Schema: EnvGroupResourceSchema(ctx)}
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

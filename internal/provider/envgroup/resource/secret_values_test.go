package resource

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		wantError bool
		wantEmpty bool
	}{
		{name: "both omitted", body: `{"id":"evg-test","name":"remote"}`, wantError: true},
		{name: "env vars omitted", body: `{"id":"evg-test","name":"remote","secretFiles":[]}`, wantError: true},
		{name: "secret files omitted", body: `{"id":"evg-test","name":"remote","envVars":[]}`, wantError: true},
		{name: "null collections", body: `{"id":"evg-test","name":"remote","envVars":null,"secretFiles":null}`, wantError: true},
		{name: "empty collections", body: `{"id":"evg-test","name":"remote","envVars":[],"secretFiles":[]}`, wantEmpty: true},
		{name: "populated collections", body: `{"id":"evg-test","name":"remote","envVars":[{"key":"TOKEN","value":"remote-token"}],"secretFiles":[{"name":"credentials","content":"remote-content"}]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &envGroupResource{client: secretValuesClient(t, http.MethodGet, tt.body)}
			state := secretValuesState(t)
			resp := resource.ReadResponse{State: state}

			r.Read(ctx, resource.ReadRequest{State: state}, &resp)

			if tt.wantError {
				require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				assert.Contains(t, resp.Diagnostics[0].Detail(), "permission to read")
				assert.True(t, state.Raw.Equal(resp.State.Raw), "unavailable secrets must not replace prior state")
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got envgroup.EnvGroupModel
			require.False(t, resp.State.Get(ctx, &got).HasError())
			assert.Equal(t, "remote", got.Name.ValueString())
			if tt.wantEmpty {
				assert.Empty(t, got.EnvVars)
				assert.Empty(t, got.SecretFiles)
				return
			}
			assert.Equal(t, map[string]common.EnvVarModel{
				"TOKEN": {Value: types.StringValue("remote-token"), GenerateValue: types.BoolValue(true)},
			}, got.EnvVars)
			assert.Equal(t, map[string]common.SecretFileModel{
				"credentials": {Content: "remote-content"},
			}, got.SecretFiles)
		})
	}
}

func TestCreateUnavailableSecretValuesRetainsID(t *testing.T) {
	ctx := context.Background()
	r := &envGroupResource{client: secretValuesClient(t, http.MethodPost, `{"id":"evg-created","name":"created"}`), ownerID: "tea-test"}
	state := secretValuesState(t)
	plan := tfsdk.Plan(state)
	resp := resource.CreateResponse{State: tfsdk.State{
		Schema: state.Schema,
		Raw:    tftypes.NewValue(state.Schema.Type().TerraformType(ctx), nil),
	}}

	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Contains(t, resp.Diagnostics[0].Detail(), "permission to read")
	var got envgroup.EnvGroupModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	assert.Equal(t, "evg-created", got.Id.ValueString(), "successful creates must remain tracked when secrets are unavailable")
}

func TestUpdateUnavailableSecretValuesPreservesState(t *testing.T) {
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
			if rename {
				var model envgroup.EnvGroupModel
				require.False(t, state.Get(ctx, &model).HasError())
				model.Name = types.StringValue("remote")
				require.False(t, plan.Set(ctx, model).HasError())
			}
			resp := resource.UpdateResponse{State: state}

			r.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, &resp)

			require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			assert.Contains(t, resp.Diagnostics[0].Detail(), "permission to read")
			assert.True(t, state.Raw.Equal(resp.State.Raw), "unavailable secrets must not replace prior state after update")
		})
	}
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

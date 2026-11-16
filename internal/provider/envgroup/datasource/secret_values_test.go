package datasource

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		{name: "empty collections", body: `{"id":"evg-test","name":"remote","envVars":[],"secretFiles":[]}`, wantEmpty: true},
		{name: "populated collections", body: `{"id":"evg-test","name":"remote","envVars":[{"key":"TOKEN","value":"remote-token"}],"secretFiles":[{"name":"credentials","content":"remote-content"}]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/env-groups/evg-test", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, err := io.WriteString(w, tt.body)
				assert.NoError(t, err)
			}))
			t.Cleanup(server.Close)
			apiClient, err := client.NewClientWithResponses(server.URL)
			require.NoError(t, err)
			d := &envGroupDataSource{client: apiClient}
			state := tfsdk.State{Schema: EnvGroupDataSourceSchema(ctx)}
			require.False(t, state.Set(ctx, envgroup.EnvGroupModel{Id: types.StringValue("evg-test")}).HasError())
			req := datasource.ReadRequest{Config: tfsdk.Config(state)}
			resp := datasource.ReadResponse{State: state}

			d.Read(ctx, req, &resp)

			if tt.wantError {
				require.True(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				assert.Contains(t, resp.Diagnostics[0].Detail(), "permission to read")
				assert.True(t, state.Raw.Equal(resp.State.Raw), "data source must not publish unavailable secrets as empty values")
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
				"TOKEN": {Value: types.StringValue("remote-token"), GenerateValue: types.BoolValue(false)},
			}, got.EnvVars)
			assert.Equal(t, map[string]common.SecretFileModel{
				"credentials": {Content: "remote-content"},
			}, got.SecretFiles)
		})
	}
}

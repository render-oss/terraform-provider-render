package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/envgroup"
)

func TestEnvGroupLinkRead(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		removed bool
		wantErr bool
	}{
		{name: "present", status: http.StatusOK, body: `{"id":"evg-example","serviceLinks":[{"id":"srv-current"}]}`},
		{name: "deleted group", status: http.StatusNotFound, body: `{"message":"not found"}`, removed: true},
		{name: "forbidden", status: http.StatusForbidden, body: `{"message":"access denied"}`, wantErr: true},
		{name: "server failure", status: http.StatusInternalServerError, body: `{"message":"unavailable"}`, wantErr: true},
		{name: "invalid response", status: http.StatusOK, body: `{`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/env-groups/evg-example", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			api, err := client.NewClientWithResponses(server.URL, client.WithHTTPClient(server.Client()))
			require.NoError(t, err)

			ctx := context.Background()
			state := tfsdk.State{Schema: EnvGroupLinkResourceSchema(ctx)}
			model := envgroup.EnvGroupLinkModel{
				EnvGroupId: types.StringValue("evg-example"),
				ServiceIds: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("srv-previous")}),
			}
			require.False(t, state.Set(ctx, &model).HasError())
			response := resource.ReadResponse{State: state}
			(&envGroupLinkResource{client: api}).Read(ctx, resource.ReadRequest{State: state}, &response)

			require.Equal(t, tc.wantErr, response.Diagnostics.HasError(), response.Diagnostics)
			if tc.removed {
				require.True(t, response.State.Raw.IsNull())
				require.Equal(t, 1, response.Diagnostics.WarningsCount())
				return
			}
			if tc.wantErr {
				require.True(t, state.Raw.Equal(response.State.Raw), "failed reads must preserve state")
				return
			}
			var actual envgroup.EnvGroupLinkModel
			require.False(t, response.State.Get(ctx, &actual).HasError())
			require.Equal(t, model.EnvGroupId, actual.EnvGroupId)
			require.True(t, actual.ServiceIds.Equal(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("srv-current")})))
		})
	}
}

func TestEnvGroupLinkCreateMissingGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	t.Cleanup(server.Close)
	api, err := client.NewClientWithResponses(server.URL, client.WithHTTPClient(server.Client()))
	require.NoError(t, err)
	ctx := context.Background()
	schema := EnvGroupLinkResourceSchema(ctx)
	plan := tfsdk.Plan{Schema: schema}
	require.False(t, plan.Set(ctx, &envgroup.EnvGroupLinkModel{
		EnvGroupId: types.StringValue("evg-example"),
		ServiceIds: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("srv-example")}),
	}).HasError())
	response := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	(&envGroupLinkResource{client: api}).Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError(), "a missing group cannot be linked during create")
}

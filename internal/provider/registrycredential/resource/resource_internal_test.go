package resource

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
)

func TestReadRemovesResourceOnNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, http.MethodGet, req.Method)
		require.Equal(t, "/registrycredentials/rgc-gone", req.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	t.Cleanup(srv.Close)

	c, err := client.NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	r := &registryCredentialResource{client: c, ownerID: "own-test"}

	ctx := t.Context()
	s := RegistryCredentialResourceSchema(ctx)

	state := tfsdk.State{Schema: s}
	require.False(t, state.Set(ctx, &RegistryCredentialModel{
		Id:        types.StringValue("rgc-gone"),
		AuthToken: types.StringValue("token"),
		Name:      types.StringValue("cred"),
		Registry:  types.StringValue("GITHUB"),
		Username:  types.StringValue("user"),
	}).HasError())

	resp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.True(t, resp.State.Raw.IsNull(), "a registry credential deleted outside Terraform must be removed from state")
	require.Len(t, resp.Diagnostics.Warnings(), 1)
	require.Equal(t, diag.SeverityWarning, resp.Diagnostics[0].Severity())
}

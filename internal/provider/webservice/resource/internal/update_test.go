package internal

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/provider/common"
	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/webservice"
)

func TestUpdateServiceRequestFromModel_PreservesDigestAfterUnrelatedUpdate(t *testing.T) {
	const (
		repository = "ghcr.io/acme/app"
		digestHex  = "1fe6fa21bcb65be49241d28f919392179db011fadf8a77ee3bfb2c314924dd4c"
		digest     = "sha256:" + digestHex
	)

	plan := webservice.WebServiceModel{
		Name: types.StringValue("unrelated-name-update"),
		Plan: types.StringValue("starter"),
		RuntimeSource: &common.RuntimeSourceModel{
			Image: &common.ImageRuntimeSourceModel{
				ImageURL:             commontypes.ImageURLStringValue{StringValue: types.StringValue(repository)},
				Tag:                  types.StringValue(digestHex),
				Digest:               types.StringValue(digest),
				RegistryCredentialID: types.StringValue("rgc-123"),
			},
		},
		PullRequestPreviewsEnabled: types.BoolValue(false),
		PreDeployCommand:           types.StringNull(),
		StartCommand:               types.StringNull(),
		MaintenanceMode:            common.DefaultMaintenanceMode(),
		Previews:                   common.PreviewsToPreviewsObject(nil),
		IPAllowList:                common.IPAllowListFromClient(nil, nil),
	}

	req, err := UpdateServiceRequestFromModel(
		context.Background(),
		plan,
		webservice.WebServiceModel{},
		"owner-123",
	)
	require.NoError(t, err)
	require.NotNil(t, req.Image)

	assert.Equal(t, repository+"@"+digest, req.Image.ImagePath)
	assert.NotEqual(t, repository+":"+digestHex, req.Image.ImagePath)
	assert.Equal(t, "rgc-123", *req.Image.RegistryCredentialId)
}

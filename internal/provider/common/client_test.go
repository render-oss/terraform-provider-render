package common_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/disks"
	"terraform-provider-render/internal/provider/common"
	th "terraform-provider-render/internal/provider/testhelpers"
)

func TestDelete(t *testing.T) {
	t.Run("it is successful when a 2xx is returned", func(t *testing.T) {
		err := common.Delete(func() (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNoContent}, nil
		})
		require.NoError(t, err)
	})
	t.Run("it is successful when a not found is returned", func(t *testing.T) {
		err := common.Delete(func() (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNotFound}, nil
		})
		require.NoError(t, err)
	})
	t.Run("it returns an error when a non-200, non-404 status is returned", func(t *testing.T) {
		err := common.Delete(func() (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusInternalServerError}, nil
		})
		require.Error(t, err)
	})
}

func TestGetWrappedService(t *testing.T) {
	t.Run("it adds env-vars", func(t *testing.T) {
		s := &client.Service{Id: "some-service-id"}

		mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
			"/services/some-service-id": th.StaticResponse(s),
			"/services/some-service-id/env-vars": th.ListResponse(
				[]client.EnvVarWithCursor{
					{EnvVar: client.EnvVar{Key: "key1", Value: "val1"}},
				},
				[]client.EnvVarWithCursor{
					{EnvVar: client.EnvVar{Key: "key2", Value: "val2"}},
				},
			),
			"/services/some-service-id/secret-files":                    th.StaticResponse([]struct{}{}),
			"/notification-settings/overrides/services/some-service-id": th.StaticResponse(struct{}{}),
		})

		c, err := client.NewClientWithResponses(mockAPI.URL)
		require.NoError(t, err)

		wrapped, err := common.GetWrappedService(context.Background(), c, "some-service-id")
		require.NoError(t, err)

		assert.Equal(t, "some-service-id", wrapped.Id)

		require.NotNil(t, wrapped.EnvVars)
		require.Len(t, *wrapped.EnvVars, 2)
		assert.Equal(t, client.EnvVar{Key: "key1", Value: "val1"}, (*wrapped.EnvVars)[0].EnvVar)
		assert.Equal(t, client.EnvVar{Key: "key2", Value: "val2"}, (*wrapped.EnvVars)[1].EnvVar)
	})

	t.Run("it adds secret files", func(t *testing.T) {
		s := &client.Service{Id: "some-service-id"}

		mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
			"/services/some-service-id":          th.StaticResponse(s),
			"/services/some-service-id/env-vars": th.StaticResponse([]struct{}{}),
			"/services/some-service-id/secret-files": th.ListResponse(
				[]client.SecretFileWithCursor{
					{SecretFile: client.SecretFile{Name: "key1", Content: "val1"}},
				},
				[]client.SecretFileWithCursor{
					{SecretFile: client.SecretFile{Name: "key2", Content: "val2"}},
				},
			),
			"/notification-settings/overrides/services/some-service-id": th.StaticResponse(struct{}{}),
		})

		c, err := client.NewClientWithResponses(mockAPI.URL)
		require.NoError(t, err)

		wrapped, err := common.GetWrappedService(context.Background(), c, "some-service-id")
		require.NoError(t, err)

		assert.Equal(t, "some-service-id", wrapped.Id)

		require.NotNil(t, wrapped.SecretFiles)
		require.Len(t, *wrapped.SecretFiles, 2)
		assert.Equal(t, client.SecretFile{Name: "key1", Content: "val1"}, (*wrapped.SecretFiles)[0].SecretFile)
		assert.Equal(t, client.SecretFile{Name: "key2", Content: "val2"}, (*wrapped.SecretFiles)[1].SecretFile)
	})
}

func TestUpdateService(t *testing.T) {
	t.Run("it updates the service", func(t *testing.T) {
		var deployCalled bool

		mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
			"/services/some-service-id": th.StaticResponse(&client.Service{
				Id: "some-service-id", Name: "updated-service",
			}),
			"/services/some-service-id/env-vars": th.StaticResponse([]client.EnvVarWithCursor{
				{EnvVar: client.EnvVar{Key: "updated-env-var", Value: "val"}},
			}),
			"/services/some-service-id/secret-files": th.ListResponse(
				[]client.SecretFileWithCursor{
					{SecretFile: client.SecretFile{Name: "updated-secret-file", Content: "val1"}},
				},
			),
			"/disks/some-disk-id": th.StaticResponse(disks.DiskDetails{Name: "updated-disk"}),
			"/services/some-service-id/deploys": func(resp http.ResponseWriter, req *http.Request) {
				deployCalled = true
				resp.WriteHeader(http.StatusCreated)
			},
			"/services/some-service-id/scale": func(resp http.ResponseWriter, req *http.Request) {
				resp.WriteHeader(http.StatusAccepted)
			},
			"/notification-settings/overrides/services/some-service-id": th.StaticResponse(struct{}{}),
		})

		c, err := client.NewClientWithResponses(mockAPI.URL)
		require.NoError(t, err)

		wrapped, err := common.UpdateService(context.Background(), c, common.UpdateServiceReq{
			ServiceID: "some-service-id",
			Disk: &common.DiskStateAndPlan{
				State: &common.DiskModel{ID: types.StringValue("some-disk-id")},
				Plan:  &common.DiskModel{ID: types.StringValue("some-disk-id")},
			},
		}, common.ServiceTypeWebService)
		require.NoError(t, err)

		assert.Equal(t, "some-service-id", wrapped.Id)

		require.Len(t, *wrapped.EnvVars, 1)
		require.Len(t, *wrapped.SecretFiles, 1)

		assert.Equal(t, "updated-env-var", (*wrapped.EnvVars)[0].EnvVar.Key)
		assert.Equal(t, "updated-secret-file", (*wrapped.SecretFiles)[0].SecretFile.Name)

		details, err := wrapped.Service.ServiceDetails.AsWebServiceDetails()
		require.NoError(t, err)

		require.NotNil(t, details.Disk)
		assert.Equal(t, "updated-disk", details.Disk.Name)

		assert.True(t, deployCalled, "it should deploy the service")
	})
}

package common_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
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
		var deployCalls atomic.Int32

		mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
			"/services/some-service-id": th.StaticResponse(&client.Service{
				Id:        "some-service-id",
				Name:      "updated-service",
				Suspended: client.ServiceSuspendedNotSuspended,
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
				deployCalls.Add(1)
				resp.WriteHeader(http.StatusCreated)
			},
			"/services/some-service-id/scale": func(resp http.ResponseWriter, req *http.Request) {
				resp.WriteHeader(http.StatusAccepted)
			},
			"/notification-settings/overrides/services/some-service-id": th.StaticResponse(struct{}{}),
		})
		t.Cleanup(mockAPI.Close)

		c, err := client.NewClientWithResponses(mockAPI.URL)
		require.NoError(t, err)

		wrapped, err := common.UpdateService(context.Background(), c, false, common.UpdateServiceReq{
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

		details, err := wrapped.ServiceDetails.AsWebServiceDetails()
		require.NoError(t, err)

		require.NotNil(t, details.Disk)
		assert.Equal(t, "updated-disk", details.Disk.Name)

		assert.EqualValues(t, 1, deployCalls.Load(), "an active service should deploy exactly once")
	})
	t.Run("it skips deploy when skip is set", func(t *testing.T) {
		var deployCalls atomic.Int32

		mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
			"/services/some-service-id": th.StaticResponse(&client.Service{
				Id:        "some-service-id",
				Name:      "updated-service",
				Suspended: client.ServiceSuspendedNotSuspended,
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
				deployCalls.Add(1)
				resp.WriteHeader(http.StatusCreated)
			},
			"/services/some-service-id/scale": func(resp http.ResponseWriter, req *http.Request) {
				resp.WriteHeader(http.StatusAccepted)
			},
			"/notification-settings/overrides/services/some-service-id": th.StaticResponse(struct{}{}),
		})
		t.Cleanup(mockAPI.Close)

		c, err := client.NewClientWithResponses(mockAPI.URL)
		require.NoError(t, err)

		_, err = common.UpdateService(context.Background(), c, true, common.UpdateServiceReq{
			ServiceID: "some-service-id",
			Disk: &common.DiskStateAndPlan{
				State: &common.DiskModel{ID: types.StringValue("some-disk-id")},
				Plan:  &common.DiskModel{ID: types.StringValue("some-disk-id")},
			},
		}, common.ServiceTypeWebService)
		require.NoError(t, err)

		assert.Zero(t, deployCalls.Load(), "an explicit skip should remain authoritative")
	})
	t.Run("it returns a deploy error for an active service", func(t *testing.T) {
		const serviceID = "active-service-id"

		var serviceUpdateCalls atomic.Int32
		var envVarUpdateCalls atomic.Int32
		var secretFileUpdateCalls atomic.Int32
		var deployCalls atomic.Int32

		mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
			"/services/" + serviceID: func(resp http.ResponseWriter, req *http.Request) {
				serviceUpdateCalls.Add(1)
				th.StaticResponse(&client.Service{
					Id:        serviceID,
					Name:      "updated-active-service",
					Suspended: client.ServiceSuspendedNotSuspended,
					Type:      client.BackgroundWorker,
				})(resp, req)
			},
			"/services/" + serviceID + "/env-vars": func(resp http.ResponseWriter, req *http.Request) {
				envVarUpdateCalls.Add(1)
				th.StaticResponse([]client.EnvVarWithCursor{})(resp, req)
			},
			"/services/" + serviceID + "/secret-files": func(resp http.ResponseWriter, req *http.Request) {
				secretFileUpdateCalls.Add(1)
				th.StaticResponse([]client.SecretFileWithCursor{})(resp, req)
			},
			"/services/" + serviceID + "/deploys": func(resp http.ResponseWriter, req *http.Request) {
				deployCalls.Add(1)
				resp.WriteHeader(http.StatusInternalServerError)
			},
		})
		t.Cleanup(mockAPI.Close)

		c, err := client.NewClientWithResponses(mockAPI.URL)
		require.NoError(t, err)

		_, err = common.UpdateService(context.Background(), c, false, common.UpdateServiceReq{
			ServiceID: serviceID,
		}, common.ServiceTypeBackgroundWorker)
		require.EqualError(t, err, "unable to deploy service: unexpected status code: 500")

		assert.EqualValues(t, 1, serviceUpdateCalls.Load(), "the service mutation must not be retried")
		assert.EqualValues(t, 1, envVarUpdateCalls.Load(), "the env-var mutation must not be retried")
		assert.EqualValues(t, 1, secretFileUpdateCalls.Load(), "the secret-file mutation must not be retried")
		assert.EqualValues(t, 1, deployCalls.Load(), "an active service should attempt exactly one deploy")
	})
}

func TestUpdateService_SuspendedServiceSkipsDeployAndReturnsUpdatedState(t *testing.T) {
	const serviceID = "suspended-service-id"

	var serviceUpdateCalls atomic.Int32
	var envVarUpdateCalls atomic.Int32
	var secretFileUpdateCalls atomic.Int32
	var deployCalls atomic.Int32

	updatedName := "updated-suspended-service"
	envVar := client.EnvVarInput{}
	require.NoError(t, envVar.FromEnvVarKeyValue(client.EnvVarKeyValue{
		Key:   "UPDATED_ENV_VAR",
		Value: "updated-value",
	}))

	mockAPI := th.NewMockRenderAPI(map[string]http.HandlerFunc{
		"/services/" + serviceID: func(resp http.ResponseWriter, req *http.Request) {
			serviceUpdateCalls.Add(1)
			assert.Equal(t, http.MethodPatch, req.Method)

			var body client.UpdateServiceJSONRequestBody
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Errorf("decode service update body: %v", err)
				http.Error(resp, "invalid service update body", http.StatusBadRequest)
				return
			}
			if body.Name == nil {
				t.Error("service update omitted name")
				http.Error(resp, "missing service name", http.StatusBadRequest)
				return
			}
			assert.Equal(t, updatedName, *body.Name)

			th.StaticResponse(&client.Service{
				Id:        serviceID,
				Name:      updatedName,
				Suspended: client.ServiceSuspendedSuspended,
				Type:      client.BackgroundWorker,
			})(resp, req)
		},
		"/services/" + serviceID + "/env-vars": func(resp http.ResponseWriter, req *http.Request) {
			envVarUpdateCalls.Add(1)
			assert.Equal(t, http.MethodPut, req.Method)
			th.StaticResponse([]client.EnvVarWithCursor{{
				EnvVar: client.EnvVar{Key: "UPDATED_ENV_VAR", Value: "updated-value"},
			}})(resp, req)
		},
		"/services/" + serviceID + "/secret-files": func(resp http.ResponseWriter, req *http.Request) {
			secretFileUpdateCalls.Add(1)
			assert.Equal(t, http.MethodPut, req.Method)
			th.StaticResponse([]client.SecretFileWithCursor{{
				SecretFile: client.SecretFile{Name: "updated-secret", Content: "updated-content"},
			}})(resp, req)
		},
		"/services/" + serviceID + "/deploys": func(resp http.ResponseWriter, req *http.Request) {
			deployCalls.Add(1)
			assert.Equal(t, http.MethodPost, req.Method)
			resp.WriteHeader(http.StatusBadRequest)
			_, _ = resp.Write([]byte(`{"message":"cannot deploy suspended service"}`))
		},
	})
	t.Cleanup(mockAPI.Close)

	c, err := client.NewClientWithResponses(mockAPI.URL)
	require.NoError(t, err)

	wrapped, err := common.UpdateService(context.Background(), c, false, common.UpdateServiceReq{
		ServiceID: serviceID,
		Service: client.UpdateServiceJSONRequestBody{
			Name: &updatedName,
		},
		EnvVars: []client.EnvVarInput{envVar},
		SecretFiles: []client.SecretFileInput{{
			Name:    "updated-secret",
			Content: "updated-content",
		}},
	}, common.ServiceTypeBackgroundWorker)
	require.NoError(t, err)
	require.NotNil(t, wrapped)

	assert.Equal(t, updatedName, wrapped.Name)
	assert.Equal(t, client.ServiceSuspendedSuspended, wrapped.Suspended)
	require.NotNil(t, wrapped.EnvVars)
	require.Len(t, *wrapped.EnvVars, 1)
	assert.Equal(t, "UPDATED_ENV_VAR", (*wrapped.EnvVars)[0].EnvVar.Key)
	require.NotNil(t, wrapped.SecretFiles)
	require.Len(t, *wrapped.SecretFiles, 1)
	assert.Equal(t, "updated-secret", (*wrapped.SecretFiles)[0].SecretFile.Name)

	assert.EqualValues(t, 1, serviceUpdateCalls.Load(), "the service mutation must not be retried")
	assert.EqualValues(t, 1, envVarUpdateCalls.Load(), "the env-var mutation must not be retried")
	assert.EqualValues(t, 1, secretFileUpdateCalls.Load(), "the secret-file mutation must not be retried")
	assert.Zero(t, deployCalls.Load(), "a suspended service must not be deployed")
}

package resource_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/notifications"
	"terraform-provider-render/internal/provider"
	"terraform-provider-render/internal/provider/common"
	th "terraform-provider-render/internal/provider/testhelpers"
)

const (
	suspendedUpdateServiceID   = "srv-suspended-update"
	suspendedUpdateServiceName = "suspended-update"
	suspendedUpdateOwnerID     = "tea-suspended-update"
	suspendedUpdateEnvKey      = "SHARED_VALUE"
)

const suspendedUpdateConfig = `
variable "env_value" {
  type = string
}

resource "render_background_worker" "suspended" {
  name   = "suspended-update"
  plan   = "starter"
  region = "oregon"

  runtime_source = {
    image = {
      image_url = "nginx"
      tag       = "latest"
    }
  }

  env_vars = {
    SHARED_VALUE = {
      value = var.env_value
    }
  }
}
`

func TestBackgroundWorkerResource_SuspendedUpdateConverges(t *testing.T) {
	fake := newSuspendedUpdateAPI(t)
	server := th.NewMockRenderAPI(map[string]http.HandlerFunc{
		"/.*": fake.handle,
	})
	t.Cleanup(server.Close)

	resourceName := "render_background_worker.suspended"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: suspendedUpdateProviderFactories(server),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       suspendedUpdateConfig,
				ConfigVariables: config.Variables{
					"env_value": config.StringVariable("before"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suspendedUpdateServiceName),
					resource.TestCheckResourceAttr(resourceName, "env_vars.SHARED_VALUE.value", "before"),
				),
			},
			{
				ResourceName: resourceName,
				Config:       suspendedUpdateConfig,
				ConfigVariables: config.Variables{
					"env_value": config.StringVariable("after"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suspendedUpdateServiceName),
					resource.TestCheckResourceAttr(resourceName, "env_vars.SHARED_VALUE.value", "after"),
				),
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "env_vars.SHARED_VALUE.value", "after"),
				),
			},
		},
	})

	requests := fake.requests()
	require.Equal(t, 1, requests.create, "the service create mutation must not be retried")
	require.Equal(t, 1, requests.patch, "the service update mutation must not be retried")
	require.Equal(t, 1, requests.envUpdate, "the env-var mutation must not be retried")
	require.Equal(t, 1, requests.secretUpdate, "the secret-file mutation must not be retried")
	require.Equal(t, 0, requests.deploy, "a suspended service must not be deployed")
	require.Equal(t, 0, requests.suspendOrResume, "the provider must not change suspension state")
	require.Equal(t, "after", requests.envValue, "the remote env var must contain the applied value")
}

func suspendedUpdateProviderFactories(server *httptest.Server) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(provider.New(
			"test",
			provider.WithHost(server.URL),
			provider.WithAPIKey("suspended-update-api-key"),
			provider.WithOwnerID(suspendedUpdateOwnerID),
			provider.WithHTTPClient(server.Client()),
			provider.WithPoller(&common.TestPoller),
			provider.WithWaitForDeployCompletion(false),
		)()),
	}
}

type suspendedUpdateRequestCounts struct {
	create          int
	patch           int
	envUpdate       int
	secretUpdate    int
	deploy          int
	suspendOrResume int
	envValue        string
}

type suspendedUpdateAPI struct {
	t *testing.T

	mu             sync.Mutex
	created        bool
	deleted        bool
	createCount    int
	patchCount     int
	envUpdateCount int
	secretUpdate   int
	deployCount    int
	lifecycleCount int
	envValue       string
}

func newSuspendedUpdateAPI(t *testing.T) *suspendedUpdateAPI {
	t.Helper()
	return &suspendedUpdateAPI{t: t}
}

func (f *suspendedUpdateAPI) handle(resp http.ResponseWriter, req *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if got := req.Header.Get("Authorization"); got != "Bearer suspended-update-api-key" {
		f.failRequest(resp, req, "unexpected Authorization header: %q", got)
		return
	}

	servicePath := "/services/" + suspendedUpdateServiceID
	switch {
	case req.Method == http.MethodPost && req.URL.Path == "/services":
		f.createService(resp, req)
	case req.Method == http.MethodGet && req.URL.Path == "/services":
		f.listServices(resp)
	case req.Method == http.MethodGet && req.URL.Path == servicePath:
		f.retrieveService(resp)
	case req.Method == http.MethodPatch && req.URL.Path == servicePath:
		f.updateService(resp, req)
	case req.Method == http.MethodDelete && req.URL.Path == servicePath:
		f.deleted = true
		resp.WriteHeader(http.StatusNoContent)
	case req.Method == http.MethodGet && req.URL.Path == servicePath+"/env-vars":
		f.getEnvVars(resp, req)
	case req.Method == http.MethodPut && req.URL.Path == servicePath+"/env-vars":
		f.updateEnvVars(resp, req)
	case req.Method == http.MethodGet && req.URL.Path == servicePath+"/secret-files":
		writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, []client.SecretFileWithCursor{})
	case req.Method == http.MethodPut && req.URL.Path == servicePath+"/secret-files":
		f.secretUpdate++
		writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, []client.SecretFileWithCursor{})
	case req.Method == http.MethodPost && req.URL.Path == servicePath+"/deploys":
		f.deployCount++
		writeSuspendedUpdateJSON(f.t, resp, http.StatusBadRequest, map[string]string{"message": "cannot deploy suspended service"})
	case req.Method == http.MethodPost && (req.URL.Path == servicePath+"/suspend" || req.URL.Path == servicePath+"/resume"):
		f.lifecycleCount++
		f.failRequest(resp, req, "provider changed suspension state")
	case req.Method == http.MethodPatch && req.URL.Path == "/notification-settings/overrides/services/"+suspendedUpdateServiceID:
		writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, notifications.NotificationOverride{
			NotificationsToSend:         notifications.NotifyOverrideDefault,
			PreviewNotificationsEnabled: notifications.NotifyPreviewOverrideDefault,
			ServiceId:                   suspendedUpdateServiceID,
		})
	case req.Method == http.MethodGet && req.URL.Path == "/notification-settings/overrides/services/"+suspendedUpdateServiceID:
		writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, notifications.NotificationOverride{
			NotificationsToSend:         notifications.NotifyOverrideDefault,
			PreviewNotificationsEnabled: notifications.NotifyPreviewOverrideDefault,
			ServiceId:                   suspendedUpdateServiceID,
		})
	case req.Method == http.MethodGet && req.URL.Path == "/logs/streams/resource/"+suspendedUpdateServiceID:
		writeSuspendedUpdateJSON(f.t, resp, http.StatusNotFound, map[string]string{"message": "not found"})
	default:
		f.failRequest(resp, req, "unexpected request")
	}
}

func (f *suspendedUpdateAPI) createService(resp http.ResponseWriter, req *http.Request) {
	var body client.CreateServiceJSONRequestBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		f.failRequest(resp, req, "decode create body: %v", err)
		return
	}

	f.createCount++
	f.created = true
	if body.OwnerId != suspendedUpdateOwnerID {
		f.t.Errorf("create owner ID = %q, want %q", body.OwnerId, suspendedUpdateOwnerID)
	}
	if body.EnvVars == nil {
		f.t.Error("create request omitted env vars")
	} else if value, err := suspendedUpdateEnvValue(*body.EnvVars); err != nil {
		f.t.Error(err)
	} else {
		f.envValue = value
	}

	writeSuspendedUpdateJSON(f.t, resp, http.StatusCreated, map[string]any{
		"deployId": "dep-suspended-update-create",
		"service":  f.service(),
	})
}

func (f *suspendedUpdateAPI) listServices(resp http.ResponseWriter) {
	if f.deleted || !f.created {
		writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, []any{})
		return
	}

	writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, []map[string]any{{
		"cursor":  "cursor-suspended-update",
		"service": f.service(),
	}})
}

func (f *suspendedUpdateAPI) retrieveService(resp http.ResponseWriter) {
	if f.deleted || !f.created {
		writeSuspendedUpdateJSON(f.t, resp, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}

	writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, f.service())
}

func (f *suspendedUpdateAPI) updateService(resp http.ResponseWriter, req *http.Request) {
	var body client.UpdateServiceJSONRequestBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		f.failRequest(resp, req, "decode update body: %v", err)
		return
	}

	f.patchCount++
	writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, f.service())
}

func (f *suspendedUpdateAPI) updateEnvVars(resp http.ResponseWriter, req *http.Request) {
	var body client.EnvVarInputArray
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		f.failRequest(resp, req, "decode env-var update body: %v", err)
		return
	}

	f.envUpdateCount++
	value, err := suspendedUpdateEnvValue(body)
	if err != nil {
		f.failRequest(resp, req, "%v", err)
		return
	}
	f.envValue = value
	writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, f.envVars())
}

func (f *suspendedUpdateAPI) getEnvVars(resp http.ResponseWriter, req *http.Request) {
	if _, hasCursor := req.URL.Query()["cursor"]; hasCursor {
		writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, []client.EnvVarWithCursor{})
		return
	}

	writeSuspendedUpdateJSON(f.t, resp, http.StatusOK, f.envVars())
}

func suspendedUpdateEnvValue(envVars client.EnvVarInputArray) (string, error) {
	if len(envVars) != 1 {
		return "", fmt.Errorf("env-var request has %d items, want 1", len(envVars))
	}

	envVar, err := envVars[0].AsEnvVarKeyValue()
	if err != nil {
		return "", fmt.Errorf("decode env-var request: %w", err)
	}
	if envVar.Key != suspendedUpdateEnvKey {
		return "", fmt.Errorf("env-var key = %q, want %q", envVar.Key, suspendedUpdateEnvKey)
	}
	return envVar.Value, nil
}

func (f *suspendedUpdateAPI) envVars() []client.EnvVarWithCursor {
	return []client.EnvVarWithCursor{{
		Cursor: "cursor-suspended-update-env",
		EnvVar: client.EnvVar{Key: suspendedUpdateEnvKey, Value: f.envValue},
	}}
}

func (f *suspendedUpdateAPI) service() client.Service {
	imagePath := "nginx:latest"
	pullRequestPreviews := client.PullRequestPreviewsEnabledNo
	previewsGeneration := client.PreviewsGenerationOff
	fixedTime := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	envSpecificDetails := client.EnvSpecificDetails{}
	if err := envSpecificDetails.FromDockerDetails(client.DockerDetails{}); err != nil {
		f.t.Errorf("encode image runtime response details: %v", err)
	}

	serviceDetails := client.Service_ServiceDetails{}
	if err := serviceDetails.FromBackgroundWorkerDetails(client.BackgroundWorkerDetails{
		BuildPlan:                  client.BuildPlanStarter,
		Env:                        client.ServiceEnvImage,
		EnvSpecificDetails:         envSpecificDetails,
		NumInstances:               1,
		Plan:                       client.PlanStarter,
		Previews:                   &client.Previews{Generation: &previewsGeneration},
		PullRequestPreviewsEnabled: &pullRequestPreviews,
		Region:                     client.Oregon,
		Runtime:                    client.ServiceRuntimeImage,
	}); err != nil {
		f.t.Errorf("encode background worker response details: %v", err)
	}

	return client.Service{
		AutoDeploy:     client.AutoDeployYes,
		CreatedAt:      fixedTime,
		DashboardUrl:   "https://dashboard.example.com/worker/" + suspendedUpdateServiceID,
		Id:             suspendedUpdateServiceID,
		ImagePath:      &imagePath,
		Name:           suspendedUpdateServiceName,
		NotifyOnFail:   client.Default,
		OwnerId:        suspendedUpdateOwnerID,
		RootDir:        "",
		ServiceDetails: serviceDetails,
		Slug:           suspendedUpdateServiceName,
		Suspended:      client.ServiceSuspendedSuspended,
		Suspenders:     []client.SuspenderType{client.SuspenderTypeUser},
		Type:           client.BackgroundWorker,
		UpdatedAt:      fixedTime,
	}
}

func (f *suspendedUpdateAPI) failRequest(resp http.ResponseWriter, req *http.Request, format string, args ...any) {
	f.t.Helper()
	f.t.Errorf("%s %s: "+format, append([]any{req.Method, req.URL.RequestURI()}, args...)...)
	writeSuspendedUpdateJSON(f.t, resp, http.StatusInternalServerError, map[string]string{"message": "unexpected request"})
}

func (f *suspendedUpdateAPI) requests() suspendedUpdateRequestCounts {
	f.mu.Lock()
	defer f.mu.Unlock()

	return suspendedUpdateRequestCounts{
		create:          f.createCount,
		patch:           f.patchCount,
		envUpdate:       f.envUpdateCount,
		secretUpdate:    f.secretUpdate,
		deploy:          f.deployCount,
		suspendOrResume: f.lifecycleCount,
		envValue:        f.envValue,
	}
}

func writeSuspendedUpdateJSON(t *testing.T, resp http.ResponseWriter, status int, body any) {
	t.Helper()
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(status)
	if err := json.NewEncoder(resp).Encode(body); err != nil {
		t.Errorf("encode fake API response: %v", err)
	}
}

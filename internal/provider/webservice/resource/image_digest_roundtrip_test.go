package resource_test

import (
	"encoding/json"
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
	digestPinnedServiceID            = "srv-digest-roundtrip"
	digestPinnedServiceName          = "digest-roundtrip"
	digestPinnedOwnerID              = "tea-digest-roundtrip"
	digestPinnedRegistryCredentialID = "rgc-digest-roundtrip"
	digestPinnedRepository           = "ghcr.io/acme/app"
	digestPinnedDigest               = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	digestPinnedImageReference       = digestPinnedRepository + "@" + digestPinnedDigest
	digestPinnedStartCommand         = "./bin/serve --port 10000"
)

const digestPinnedConfig = `
variable "start_command" {
  type    = string
  default = null
}

resource "render_web_service" "digest" {
  name           = "digest-roundtrip"
  plan           = "starter"
  region         = "oregon"
  root_directory = "apps/api"
  start_command  = var.start_command

  runtime_source = {
    image = {
      image_url              = "ghcr.io/acme/app"
      digest                 = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
      registry_credential_id = "rgc-digest-roundtrip"
    }
  }
}
`

func TestWebServiceResource_DigestPinnedImageRoundTrip(t *testing.T) {
	fake := newDigestPinnedWebServiceAPI(t)
	server := th.NewMockRenderAPI(map[string]http.HandlerFunc{
		"/.*": fake.handle,
	})
	t.Cleanup(server.Close)

	resourceName := "render_web_service.digest"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: digestPinnedProviderFactories(server),
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       digestPinnedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", digestPinnedRepository),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.digest", digestPinnedDigest),
					resource.TestCheckNoResourceAttr(resourceName, "runtime_source.image.tag"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.registry_credential_id", digestPinnedRegistryCredentialID),
					resource.TestCheckResourceAttr(resourceName, "root_directory", "apps/api"),
				),
			},
			{
				ResourceName: resourceName,
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", digestPinnedRepository),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.digest", digestPinnedDigest),
					resource.TestCheckNoResourceAttr(resourceName, "runtime_source.image.tag"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.registry_credential_id", digestPinnedRegistryCredentialID),
					resource.TestCheckResourceAttr(resourceName, "root_directory", "apps/api"),
				),
			},
			{
				ResourceName: resourceName,
				Config:       digestPinnedConfig,
				ConfigVariables: config.Variables{
					"start_command": config.StringVariable(digestPinnedStartCommand),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "start_command", digestPinnedStartCommand),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.image_url", digestPinnedRepository),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.digest", digestPinnedDigest),
					resource.TestCheckNoResourceAttr(resourceName, "runtime_source.image.tag"),
					resource.TestCheckResourceAttr(resourceName, "runtime_source.image.registry_credential_id", digestPinnedRegistryCredentialID),
					resource.TestCheckResourceAttr(resourceName, "root_directory", "apps/api"),
					resource.TestCheckResourceAttr(resourceName, "plan", "starter"),
					resource.TestCheckResourceAttr(resourceName, "region", "oregon"),
				),
			},
		},
	})

	createCount, patchCount, patchImageReferences, patchRegistryCredentialIDs := fake.requests()
	require.Equal(t, 1, createCount, "the service create mutation must not be retried")
	require.Equal(t, 1, patchCount, "the unrelated update should issue one service PATCH")
	require.Equal(t, []string{digestPinnedImageReference}, patchImageReferences)
	require.Equal(t, []string{digestPinnedRegistryCredentialID}, patchRegistryCredentialIDs)
}

func digestPinnedProviderFactories(server *httptest.Server) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(provider.New(
			"test",
			provider.WithHost(server.URL),
			provider.WithAPIKey("digest-roundtrip-api-key"),
			provider.WithOwnerID(digestPinnedOwnerID),
			provider.WithHTTPClient(server.Client()),
			provider.WithPoller(&common.TestPoller),
			provider.WithWaitForDeployCompletion(false),
		)()),
	}
}

type digestPinnedWebServiceAPI struct {
	t *testing.T

	mu                         sync.Mutex
	created                    bool
	deleted                    bool
	createCount                int
	patchCount                 int
	patchImageReferences       []string
	patchRegistryCredentialIDs []string
	name                       string
	rootDirectory              string
	startCommand               string
}

func newDigestPinnedWebServiceAPI(t *testing.T) *digestPinnedWebServiceAPI {
	t.Helper()

	return &digestPinnedWebServiceAPI{
		t:             t,
		name:          digestPinnedServiceName,
		rootDirectory: "apps/api",
	}
}

func (f *digestPinnedWebServiceAPI) handle(resp http.ResponseWriter, req *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if got := req.Header.Get("Authorization"); got != "Bearer digest-roundtrip-api-key" {
		f.failRequest(resp, req, "unexpected Authorization header: %q", got)
		return
	}

	servicePath := "/services/" + digestPinnedServiceID
	switch {
	case req.Method == http.MethodPost && req.URL.Path == "/services":
		f.createService(resp, req)
	case req.Method == http.MethodGet && req.URL.Path == "/services":
		f.listServices(resp, req)
	case req.Method == http.MethodGet && req.URL.Path == servicePath:
		f.retrieveService(resp)
	case req.Method == http.MethodPatch && req.URL.Path == servicePath:
		f.updateService(resp, req)
	case req.Method == http.MethodDelete && req.URL.Path == servicePath:
		f.deleted = true
		resp.WriteHeader(http.StatusNoContent)
	case req.Method == http.MethodGet && req.URL.Path == servicePath+"/env-vars":
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, []client.EnvVarWithCursor{})
	case req.Method == http.MethodPut && req.URL.Path == servicePath+"/env-vars":
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, []client.EnvVarWithCursor{})
	case req.Method == http.MethodGet && req.URL.Path == servicePath+"/secret-files":
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, []client.SecretFileWithCursor{})
	case req.Method == http.MethodPut && req.URL.Path == servicePath+"/secret-files":
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, []client.SecretFileWithCursor{})
	case req.Method == http.MethodGet && req.URL.Path == servicePath+"/custom-domains":
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, []client.CustomDomainWithCursor{})
	case req.Method == http.MethodPost && req.URL.Path == servicePath+"/deploys":
		writeDigestPinnedJSON(f.t, resp, http.StatusCreated, map[string]string{"id": "dep-digest-roundtrip-update"})
	case (req.Method == http.MethodGet || req.Method == http.MethodPatch) && req.URL.Path == "/notification-settings/overrides/services/"+digestPinnedServiceID:
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, notifications.NotificationOverride{
			NotificationsToSend:         notifications.NotifyOverrideDefault,
			PreviewNotificationsEnabled: notifications.NotifyPreviewOverrideDefault,
			ServiceId:                   digestPinnedServiceID,
		})
	case req.Method == http.MethodGet && req.URL.Path == "/logs/streams/resource/"+digestPinnedServiceID:
		writeDigestPinnedJSON(f.t, resp, http.StatusNotFound, map[string]string{"message": "not found"})
	default:
		f.failRequest(resp, req, "unexpected request")
	}
}

func (f *digestPinnedWebServiceAPI) createService(resp http.ResponseWriter, req *http.Request) {
	var body client.CreateServiceJSONRequestBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		f.failRequest(resp, req, "decode create body: %v", err)
		return
	}

	f.createCount++
	f.created = true
	f.name = body.Name
	if body.RootDir != nil {
		f.rootDirectory = *body.RootDir
	}

	if body.OwnerId != digestPinnedOwnerID {
		f.t.Errorf("create owner ID = %q, want %q", body.OwnerId, digestPinnedOwnerID)
	}
	if body.Image == nil {
		f.t.Error("create request omitted image")
	} else {
		if body.Image.ImagePath != digestPinnedImageReference {
			f.t.Errorf("create image path = %q, want %q", body.Image.ImagePath, digestPinnedImageReference)
		}
		if valueOrEmpty(body.Image.RegistryCredentialId) != digestPinnedRegistryCredentialID {
			f.t.Errorf("create registry credential ID = %q, want %q", valueOrEmpty(body.Image.RegistryCredentialId), digestPinnedRegistryCredentialID)
		}
	}

	writeDigestPinnedJSON(f.t, resp, http.StatusCreated, map[string]any{
		"deployId": "dep-digest-roundtrip",
		"service":  f.service(),
	})
}

func (f *digestPinnedWebServiceAPI) listServices(resp http.ResponseWriter, req *http.Request) {
	if f.deleted || !f.created {
		writeDigestPinnedJSON(f.t, resp, http.StatusOK, []any{})
		return
	}

	if got := req.URL.Query().Get("name"); got != digestPinnedServiceName {
		f.t.Errorf("service list name = %q, want %q", got, digestPinnedServiceName)
	}
	if got := req.URL.Query().Get("ownerId"); got != digestPinnedOwnerID {
		f.t.Errorf("service list ownerId = %q, want %q", got, digestPinnedOwnerID)
	}
	if got := req.URL.Query().Get("type"); got != string(client.WebService) {
		f.t.Errorf("service list type = %q, want %q", got, client.WebService)
	}

	writeDigestPinnedJSON(f.t, resp, http.StatusOK, []map[string]any{{
		"cursor":  "cursor-digest-roundtrip",
		"service": f.service(),
	}})
}

func (f *digestPinnedWebServiceAPI) retrieveService(resp http.ResponseWriter) {
	if f.deleted || !f.created {
		writeDigestPinnedJSON(f.t, resp, http.StatusNotFound, map[string]string{"message": "not found"})
		return
	}

	writeDigestPinnedJSON(f.t, resp, http.StatusOK, f.service())
}

func (f *digestPinnedWebServiceAPI) updateService(resp http.ResponseWriter, req *http.Request) {
	var body client.UpdateServiceJSONRequestBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		f.failRequest(resp, req, "decode update body: %v", err)
		return
	}

	f.patchCount++
	if body.Name != nil {
		f.name = *body.Name
	}
	if body.RootDir != nil {
		f.rootDirectory = *body.RootDir
	}

	if body.Image == nil {
		f.t.Error("update request omitted image")
	} else {
		f.patchImageReferences = append(f.patchImageReferences, body.Image.ImagePath)
		f.patchRegistryCredentialIDs = append(f.patchRegistryCredentialIDs, valueOrEmpty(body.Image.RegistryCredentialId))
		if body.Image.ImagePath != digestPinnedImageReference {
			f.t.Errorf("update image path = %q, want %q", body.Image.ImagePath, digestPinnedImageReference)
		}
	}

	if body.ServiceDetails == nil {
		f.t.Error("update request omitted service details")
	} else {
		details, err := body.ServiceDetails.AsWebServiceDetailsPATCH()
		if err != nil {
			f.t.Errorf("decode web service update details: %v", err)
		} else if details.EnvSpecificDetails == nil {
			f.t.Error("update request omitted image runtime details")
		} else {
			dockerDetails, err := details.EnvSpecificDetails.AsDockerDetailsPATCH()
			if err != nil {
				f.t.Errorf("decode image runtime update details: %v", err)
			} else if dockerDetails.DockerCommand == nil {
				f.t.Error("update request omitted start command")
			} else {
				f.startCommand = *dockerDetails.DockerCommand
			}
		}
	}

	if f.startCommand != digestPinnedStartCommand {
		f.t.Errorf("updated start command = %q, want %q", f.startCommand, digestPinnedStartCommand)
	}

	writeDigestPinnedJSON(f.t, resp, http.StatusOK, f.service())
}

func (f *digestPinnedWebServiceAPI) service() client.Service {
	imagePath := digestPinnedImageReference
	pullRequestPreviews := client.PullRequestPreviewsEnabledNo
	previewsGeneration := client.PreviewsGenerationOff
	fixedTime := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	envSpecificDetails := client.EnvSpecificDetails{}
	if err := envSpecificDetails.FromDockerDetails(client.DockerDetails{
		DockerCommand:    f.startCommand,
		DockerContext:    "",
		DockerfilePath:   "",
		PreDeployCommand: nil,
	}); err != nil {
		f.t.Errorf("encode image runtime response details: %v", err)
	}

	serviceDetails := client.Service_ServiceDetails{}
	if err := serviceDetails.FromWebServiceDetails(client.WebServiceDetails{
		BuildPlan:                  client.BuildPlanStarter,
		Env:                        client.ServiceEnvImage,
		EnvSpecificDetails:         envSpecificDetails,
		HealthCheckPath:            "",
		MaintenanceMode:            &client.MaintenanceMode{Enabled: false, Uri: ""},
		NumInstances:               1,
		OpenPorts:                  []client.ServerPort{},
		Plan:                       client.PlanStarter,
		Previews:                   &client.Previews{Generation: &previewsGeneration},
		PullRequestPreviewsEnabled: &pullRequestPreviews,
		Region:                     client.Oregon,
		Runtime:                    client.ServiceRuntimeImage,
		Url:                        "https://digest-roundtrip.example.com",
	}); err != nil {
		f.t.Errorf("encode web service response details: %v", err)
	}

	return client.Service{
		AutoDeploy:     client.AutoDeployYes,
		CreatedAt:      fixedTime,
		DashboardUrl:   "https://dashboard.example.com/web/" + digestPinnedServiceID,
		Id:             digestPinnedServiceID,
		ImagePath:      &imagePath,
		Name:           f.name,
		NotifyOnFail:   client.Default,
		OwnerId:        digestPinnedOwnerID,
		RootDir:        f.rootDirectory,
		ServiceDetails: serviceDetails,
		Slug:           digestPinnedServiceName,
		Suspended:      client.ServiceSuspendedNotSuspended,
		Suspenders:     []client.SuspenderType{},
		Type:           client.WebService,
		UpdatedAt:      fixedTime,
		RegistryCredential: &client.RegistryCredentialSummary{
			Id:   digestPinnedRegistryCredentialID,
			Name: "digest round-trip credential",
		},
	}
}

func (f *digestPinnedWebServiceAPI) failRequest(resp http.ResponseWriter, req *http.Request, format string, args ...any) {
	f.t.Helper()
	f.t.Errorf("%s %s: "+format, append([]any{req.Method, req.URL.RequestURI()}, args...)...)
	writeDigestPinnedJSON(f.t, resp, http.StatusInternalServerError, map[string]string{"message": "unexpected request"})
}

func (f *digestPinnedWebServiceAPI) requests() (int, int, []string, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.createCount,
		f.patchCount,
		append([]string(nil), f.patchImageReferences...),
		append([]string(nil), f.patchRegistryCredentialIDs...)
}

func writeDigestPinnedJSON(t *testing.T, resp http.ResponseWriter, status int, body any) {
	t.Helper()
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(status)
	if err := json.NewEncoder(resp).Encode(body); err != nil {
		t.Errorf("encode fake API response: %v", err)
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
)

var errNotFound = fmt.Errorf("not found")

func IsNotFoundErr(err error) bool {
	return errors.Is(err, errNotFound)
}

func EmitNotFoundWarning(resourceID string, diags *diag.Diagnostics) {
	diags.AddWarning(
		"Resource not found",
		fmt.Sprintf("Resource with ID %s not found. It may have been deleted outside of Terraform. Removing object from state.",
			resourceID,
		),
	)
}

type PollCfg struct {
	MaxPollInterval      time.Duration
	StartingPollInterval time.Duration
}

var DefaultPoller = Poller{
	cfg: PollCfg{
		MaxPollInterval:      15 * time.Second,
		StartingPollInterval: 3 * time.Second,
	},
}

var TestPoller = Poller{
	cfg: PollCfg{},
}

type Poller struct {
	cfg PollCfg
}

func (p *Poller) Poll(ctx context.Context, pollFunc func() (donePolling bool, err error), timeout time.Duration) error {
	pollInterval := p.cfg.StartingPollInterval

	startTime := time.Now()

	for {
		if donePolling, err := pollFunc(); err != nil {
			return err
		} else if donePolling {
			return nil
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("timed out")
		}

		select {
		case <-time.After(pollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}

		pollInterval = time.Duration(math.Ceil(float64(pollInterval) * 1.2))
		if pollInterval > p.cfg.MaxPollInterval {
			pollInterval = p.cfg.MaxPollInterval
		}
	}
}

func Get(get func() (*http.Response, error), v any) error {
	return doForBody(get, v)
}

func Create(create func() (*http.Response, error), v any) error {
	return doForBody(create, v)
}

func Update(update func() (*http.Response, error), v any) error {
	return doForBody(update, v)
}

func Delete(del func() (*http.Response, error)) error {
	_, err := do(del)
	if errors.Is(err, errNotFound) {
		return nil
	}

	return err
}

func doForBody(f func() (*http.Response, error), v any) error {
	resp, err := do(f)
	if err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func do(f func() (*http.Response, error)) (*http.Response, error) {
	resp, err := f()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		var badRequest client.N400BadRequest
		err := json.NewDecoder(resp.Body).Decode(&badRequest)
		if err != nil {
			return nil, fmt.Errorf("bad request")
		}

		return nil, fmt.Errorf(*badRequest.Message)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		var clientError client.Error
		err := json.NewDecoder(resp.Body).Decode(&clientError)
		if err != nil {
			return nil, fmt.Errorf("received %d", resp.StatusCode)
		}

		return nil, fmt.Errorf(*clientError.Message)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return resp, nil
}

type WrappedService struct {
	*client.Service
	CustomDomains        *[]client.CustomDomain
	EnvVars              *[]client.EnvVarWithCursor
	SecretFiles          *[]client.SecretFileWithCursor
	NotificationOverride *client.NotificationOverride
}

func WrapService(ctx context.Context, apiClient *client.ClientWithResponses, service *client.Service) (*WrappedService, error) {
	envVars, err := getEnvVars(ctx, apiClient, service.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting env vars: %w", err)
	}

	secretFiles, err := getSecretFiles(ctx, apiClient, service.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting secret files: %w", err)
	}

	customDomains, err := getCustomDomains(ctx, apiClient, service)
	if err != nil {
		return nil, err
	}

	notificationOverrides, err := getNotificationOverrides(ctx, apiClient, service.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting notification overrides: %w", err)
	}

	return &WrappedService{
		Service:              service,
		CustomDomains:        customDomains,
		EnvVars:              envVars,
		SecretFiles:          secretFiles,
		NotificationOverride: notificationOverrides,
	}, nil
}

func GetWrappedService(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*WrappedService, error) {
	service, err := GetService(ctx, apiClient, serviceID)
	if err != nil {
		return nil, err
	}

	return WrapService(ctx, apiClient, service)
}

func GetWrappedServiceByName(ctx context.Context, apiClient *client.ClientWithResponses, owner, name string, serviceType client.ServiceType) (*WrappedService, error) {
	service, err := getServiceByName(ctx, apiClient, owner, name, serviceType)
	if err != nil {
		return nil, err
	}

	return WrapService(ctx, apiClient, service)
}

type serviceWithCursor struct {
	Cursor  *client.Cursor  `json:"cursor,omitempty"`
	Service *client.Service `json:"service,omitempty"`
}

func getServiceByName(ctx context.Context, apiClient *client.ClientWithResponses, owner, name string, serviceType client.ServiceType) (*client.Service, error) {
	var res []serviceWithCursor
	err := Get(func() (*http.Response, error) {
		return apiClient.GetServices(ctx, &client.GetServicesParams{
			Name:    From([]string{name}),
			OwnerId: From([]string{owner}),
			Type:    From([]client.ServiceType{serviceType}),
		})
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("could not get service: %w", err)
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("expected one service, got %d", len(res))
	}
	return res[0].Service, nil
}

func GetService(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*client.Service, error) {
	var res client.Service
	err := Get(func() (*http.Response, error) {
		return apiClient.GetService(ctx, serviceID)
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("could not get service: %w", err)
	}
	return &res, nil
}

func getEnvVars(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*[]client.EnvVarWithCursor, error) {
	var res []client.EnvVarWithCursor
	var cursor *string

	for {
		var evs []client.EnvVarWithCursor
		err := Get(func() (*http.Response, error) {
			return apiClient.GetEnvVarsForService(ctx, serviceID, &client.GetEnvVarsForServiceParams{Cursor: cursor})
		}, &evs)
		if err != nil {
			return nil, fmt.Errorf("could not get env vars for service: %w", err)
		}

		if len(evs) == 0 {
			break
		}

		cursor = &(evs[len(evs)-1].Cursor)
		res = append(res, evs...)
	}
	return &res, nil
}

func getSecretFiles(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*[]client.SecretFileWithCursor, error) {
	var res []client.SecretFileWithCursor
	var cursor *string

	for {
		var secretFiles []client.SecretFileWithCursor
		err := Get(func() (*http.Response, error) {
			return apiClient.GetSecretFilesForService(ctx, serviceID, &client.GetSecretFilesForServiceParams{Cursor: cursor})
		}, &secretFiles)
		if err != nil {
			return nil, fmt.Errorf("unable to get secret files for service: %w", err)
		}

		if len(secretFiles) == 0 {
			break
		}

		cursor = &(secretFiles[len(secretFiles)-1].Cursor)
		res = append(res, secretFiles...)
	}
	return &res, nil
}

func GetEnvironmentById(ctx context.Context, apiClient *client.ClientWithResponses, envID string) (*client.Environment, error) {
	var res client.Environment
	err := Get(func() (*http.Response, error) {
		return apiClient.GetEnvironment(ctx, envID)
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("could not get environment id %s: %w", envID, err)
	}
	return &res, nil
}

func getNotificationOverrides(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*client.NotificationOverride, error) {
	var res client.NotificationOverride
	err := Get(func() (*http.Response, error) {
		return apiClient.GetServiceNotificationOverrides(ctx, serviceID)
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("could not get notification override for service: %w", err)
	}
	return &res, nil
}

type CreateServiceReq struct {
	Service              client.CreateServiceJSONRequestBody
	CustomDomains        []client.CustomDomain
	EnvironmentID        *string
	NotificationOverride types.Object
}

type serviceWithDeploy struct {
	DeployId *string         `json:"deployId,omitempty"`
	Service  *client.Service `json:"service,omitempty"`
}

func CreateService(ctx context.Context, apiClient *client.ClientWithResponses, req CreateServiceReq) (*WrappedService, error) {
	serviceResponse := serviceWithDeploy{}
	err := Create(func() (*http.Response, error) {
		return apiClient.CreateService(ctx, req.Service)
	}, &serviceResponse)
	if err != nil {
		return nil, fmt.Errorf("could not create service: %w", err)
	}

	if req.EnvironmentID != nil {
		_, err = UpdateEnvironmentID(ctx, apiClient, serviceResponse.Service.Id, &EnvironmentIDStateAndPlan{Plan: req.EnvironmentID})
		if err != nil {
			return nil, fmt.Errorf("could not add service to environment: %w", err)
		}
		serviceResponse.Service.EnvironmentId = req.EnvironmentID
	}

	if req.CustomDomains != nil {
		err := updateCustomDomains(
			ctx,
			apiClient,
			serviceResponse.Service.Id,
			CustomDomainStateAndPlan{Plan: CustomDomainClientsToCustomDomainModels(&req.CustomDomains)},
		)
		if err != nil {
			return nil, fmt.Errorf("could not add custom domains: %w", err)
		}
	}

	notificationOverride, err := NotificationOverrideToClient(req.NotificationOverride)
	if err != nil {
		return nil, fmt.Errorf("could not process notification override: %w", err)
	}

	if notificationOverride != nil {
		err = Create(func() (*http.Response, error) {
			return apiClient.PatchServiceNotificationOverrides(ctx, serviceResponse.Service.Id, *notificationOverride)
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("could not add notification overrides: %w", err)
		}
	}

	wrappedService, err := WrapService(ctx, apiClient, serviceResponse.Service)
	if err != nil {
		return nil, fmt.Errorf("could not wrap service: %w", err)
	}

	return wrappedService, nil
}

var validCustomDomainServiceTypes = []client.ServiceType{
	client.WebService,
	client.StaticSite,
}

func getCustomDomains(ctx context.Context, apiClient *client.ClientWithResponses, service *client.Service) (*[]client.CustomDomain, error) {
	if !slices.Contains(validCustomDomainServiceTypes, service.Type) {
		return nil, nil
	}

	var res []client.CustomDomain
	var cursor *string
	limit := 100

	for {
		var cds []*client.CustomDomainWithCursor
		err := Get(func() (*http.Response, error) {
			return apiClient.GetCustomDomains(ctx, service.Id, &client.GetCustomDomainsParams{
				Cursor: cursor,
				Limit:  From(client.LimitParam(limit)),
			})
		}, &cds)
		if err != nil {
			return nil, fmt.Errorf("could not get custom domains for service: %w", err)
		}

		for _, cd := range cds {
			res = append(res, *cd.CustomDomain)
		}

		if len(cds) < limit {
			break
		}
		cursor = &(cds[len(cds)-1].Cursor)
	}

	return &res, nil
}

type DeployWithCursor struct {
	Cursor *client.Cursor `json:"cursor,omitempty"`
	Deploy *client.Deploy `json:"deploy,omitempty"`
}

func WaitForService(ctx context.Context, poller *Poller, apiClient *client.ClientWithResponses, serviceID string) error {
	return poller.Poll(ctx, func() (bool, error) {
		var deploys []DeployWithCursor
		err := Get(func() (*http.Response, error) {
			return apiClient.GetDeploys(ctx, serviceID, nil)
		}, &deploys)
		if err != nil {
			return false, err
		}

		if len(deploys) == 0 {
			return false, nil
		}

		latestDeploy := deploys[0]
		for _, deploy := range deploys {
			if latestDeploy.Deploy.CreatedAt != nil && deploy.Deploy.CreatedAt.After(*latestDeploy.Deploy.CreatedAt) {
				latestDeploy = deploy
			}
		}

		if latestDeploy.Deploy != nil && latestDeploy.Deploy.Status != nil {
			switch *latestDeploy.Deploy.Status {
			case client.Live:
				return true, nil
			case client.BuildFailed, client.Canceled, client.Deactivated, client.PreDeployFailed, client.UpdateFailed:
				return false, fmt.Errorf("deploy failed")
			case client.BuildInProgress, client.Created, client.PreDeployInProgress, client.UpdateInProgress:
				return false, nil
			}
		}

		return false, nil
	},
		// Wait up to 3 hours for the service to be live because we must wait for the build (2 hour limit),
		// pre deploy command (30 minute limit), and deploy (15 minute limit) to complete
		3*60*time.Minute,
	)
}

type UpdateServiceReq struct {
	ServiceID            string
	Service              client.UpdateServiceJSONRequestBody
	EnvironmentID        *EnvironmentIDStateAndPlan
	EnvVars              client.EnvVarInputArray
	SecretFiles          []client.SecretFileInput
	CustomDomains        CustomDomainStateAndPlan
	Disk                 *DiskStateAndPlan
	InstanceCount        *int64
	Autoscaling          *AutoscalingStateAndPlan
	NotificationOverride *client.NotificationServiceOverridePATCH
}

type AutoscalingStateAndPlan struct {
	State *AutoscalingModel
	Plan  *AutoscalingModel
}

type DiskStateAndPlan struct {
	State *DiskModel
	Plan  *DiskModel
}

type EnvironmentIDStateAndPlan struct {
	State *string
	Plan  *string
}

type CustomDomainStateAndPlan struct {
	State []CustomDomainModel
	Plan  []CustomDomainModel
}

type ServiceType string

const (
	ServiceTypeWebService       ServiceType = "web_service"
	ServiceTypePrivateService   ServiceType = "private_service"
	ServiceTypeBackgroundWorker ServiceType = "background_worker"
	ServiceTypeCronJob          ServiceType = "cron_job"
	ServiceTypeStaticSite       ServiceType = "static_site"
)

var scalableServiceTypes = []ServiceType{
	ServiceTypeWebService,
	ServiceTypePrivateService,
	ServiceTypeBackgroundWorker,
}

func UpdateService(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq, serviceType ServiceType) (*WrappedService, error) {
	// must happen before updating the service so the instance count is reflected in the service response
	if req.InstanceCount != nil && *req.InstanceCount > 0 && slices.Contains(scalableServiceTypes, serviceType) {
		if err := updateInstanceCount(ctx, apiClient, req.ServiceID, int(*req.InstanceCount)); err != nil {
			return nil, err
		}
	}

	if err := updateAutoscalingConfig(ctx, apiClient, req); err != nil {
		return nil, err
	}

	service, err := updateService(ctx, apiClient, req)
	if err != nil {
		return nil, err
	}

	envVars, err := updateEnvVars(ctx, apiClient, req)
	if err != nil {
		return nil, err
	}

	secretFiles, err := updateSecretFiles(ctx, apiClient, req)
	if err != nil {
		return nil, err
	}

	disk, err := updateDisk(ctx, apiClient, req)
	if err != nil {
		return nil, err
	}

	notificationOverride, err := updateNotificationOverride(ctx, apiClient, req)
	if err != nil {
		return nil, err
	}

	envID, err := UpdateEnvironmentID(ctx, apiClient, req.ServiceID, req.EnvironmentID)
	if err != nil {
		return nil, err
	}
	service.EnvironmentId = envID

	if err := updateCustomDomains(ctx, apiClient, req.ServiceID, req.CustomDomains); err != nil {
		return nil, err
	}

	cds, err := getCustomDomains(ctx, apiClient, service)
	if err != nil {
		return nil, fmt.Errorf("could not get custom domains: %w", err)
	}

	switch serviceType {
	case ServiceTypeWebService:
		if details, ok := service.ServiceDetails.AsWebServiceDetails(); ok == nil {
			details.Disk = DiskDetailsToDisk(disk)
			err := service.ServiceDetails.FromWebServiceDetails(details)
			if err != nil {
				return nil, fmt.Errorf("could not update service: %w", err)
			}
		}
	case ServiceTypePrivateService:
		if details, ok := service.ServiceDetails.AsPrivateServiceDetails(); ok == nil {
			details.Disk = DiskDetailsToDisk(disk)
			err := service.ServiceDetails.FromPrivateServiceDetails(details)
			if err != nil {
				return nil, fmt.Errorf("could not update service: %w", err)
			}
		}
	case ServiceTypeBackgroundWorker:
		if details, ok := service.ServiceDetails.AsBackgroundWorkerDetails(); ok == nil {
			details.Disk = DiskDetailsToDisk(disk)
			err := service.ServiceDetails.FromBackgroundWorkerDetails(details)
			if err != nil {
				return nil, fmt.Errorf("could not update service: %w", err)
			}
		}
	}

	err = Create(func() (*http.Response, error) {
		return apiClient.CreateDeploy(ctx, req.ServiceID, client.CreateDeployJSONRequestBody{})
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to deploy service: %w", err)
	}

	return &WrappedService{
		Service:       service,
		CustomDomains: cds, EnvVars: envVars,
		SecretFiles:          secretFiles,
		NotificationOverride: notificationOverride,
	}, nil
}

func updateService(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq) (*client.Service, error) {
	var res client.Service
	err := Update(func() (*http.Response, error) {
		return apiClient.UpdateService(ctx, req.ServiceID, req.Service)
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("could not update service: %w", err)
	}

	return &res, nil
}

func updateEnvVars(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq) (*[]client.EnvVarWithCursor, error) {
	var envVarResp []client.EnvVarWithCursor
	err := Update(func() (*http.Response, error) {
		return apiClient.UpdateEnvVarsForService(ctx, req.ServiceID, req.EnvVars)
	}, &envVarResp)
	if err != nil {
		return nil, fmt.Errorf("could not update env vars: %w", err)
	}

	return &envVarResp, nil
}

func updateSecretFiles(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq) (*[]client.SecretFileWithCursor, error) {
	var secretFileResp []client.SecretFileWithCursor
	err := Update(func() (*http.Response, error) {
		return apiClient.UpdateSecretFilesForService(ctx, req.ServiceID, req.SecretFiles)
	}, &secretFileResp)
	if err != nil {
		return nil, fmt.Errorf("could not update secret files: %w", err)
	}

	return &secretFileResp, nil
}

func updateDisk(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq) (*client.DiskDetails, error) {
	if req.Disk == nil || req.Disk.Plan == nil && req.Disk.State == nil {
		return nil, nil
	}

	if req.Disk.Plan == nil && req.Disk.State != nil {
		// The disk was removed
		err := Delete(func() (*http.Response, error) {
			return apiClient.DeleteDisk(ctx, req.Disk.State.ID.ValueString())
		})
		if err != nil {
			return nil, fmt.Errorf("could not delete disk: %w", err)
		}
		return nil, nil
	}

	var diskResp client.DiskDetails

	if req.Disk.Plan != nil && req.Disk.State == nil {
		// The disk was added
		err := Create(func() (*http.Response, error) {
			return apiClient.AddDisk(ctx, DiskToClientPOST(req.ServiceID, *req.Disk.Plan))
		}, &diskResp)
		if err != nil {
			return nil, fmt.Errorf("could not add disk: %w", err)
		}
		return &diskResp, nil
	}

	// The disk was updated
	err := Update(func() (*http.Response, error) {
		return apiClient.UpdateDisk(ctx, req.Disk.Plan.ID.ValueString(), DiskToClientPatch(*req.Disk.Plan))
	}, &diskResp)
	if err != nil {
		return nil, fmt.Errorf("could not update disk: %w", err)
	}
	return &diskResp, nil
}

func updateNotificationOverride(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq) (*client.NotificationOverride, error) {
	if req.NotificationOverride == nil {
		return nil, nil
	}

	var notificationOverrideResp client.NotificationOverride
	err := Update(func() (*http.Response, error) {
		return apiClient.PatchServiceNotificationOverrides(ctx, req.ServiceID, *req.NotificationOverride)
	}, &notificationOverrideResp)
	if err != nil {
		return nil, fmt.Errorf("could not update notification override: %w", err)
	}

	return &notificationOverrideResp, nil
}

func updateInstanceCount(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string, instanceCount int) error {
	err := Update(func() (*http.Response, error) {
		return apiClient.ScaleService(ctx, serviceID, client.ScaleServiceJSONRequestBody{NumInstances: instanceCount})
	}, nil)
	if err != nil {
		return fmt.Errorf("could not update instance count: %w", err)
	}
	return nil
}

func updateAutoscalingConfig(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateServiceReq) error {
	if req.Autoscaling == nil ||
		(req.Autoscaling.Plan == nil && req.Autoscaling.State == nil) {
		return nil
	}

	if req.Autoscaling.Plan == nil {
		return Delete(func() (*http.Response, error) {
			return apiClient.DeleteAutoscalingConfig(ctx, req.ServiceID)
		})
	}

	autoscaling, err := AutoscalingRequest(req.Autoscaling.Plan)
	if err != nil {
		return fmt.Errorf("could not create autoscaling config: %w", err)
	}

	return Update(func() (*http.Response, error) {
		return apiClient.AutoscaleService(ctx, req.ServiceID, *autoscaling)
	}, nil)
}

// UpdateEnvironmentID updates the environment that a resource belongs to
func UpdateEnvironmentID(ctx context.Context, apiClient *client.ClientWithResponses, resourceID string, envIDStateAndPlan *EnvironmentIDStateAndPlan) (*string, error) {
	if envIDStateAndPlan == nil {
		return nil, nil
	}

	state := envIDStateAndPlan.State
	plan := envIDStateAndPlan.Plan
	if state == plan {
		// doesn't matter which one we return
		return state, nil
	}

	if envIDStateAndPlan.Plan == nil {
		// previous was in an environment, removing
		err := Update(func() (*http.Response, error) {
			return apiClient.RemoveResourcesFromEnvironment(ctx, *state, &client.RemoveResourcesFromEnvironmentParams{
				ResourceIds: []string{resourceID},
			})
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("could not remove service from previous environment: %w", err)
		}

		return nil, nil
	} else if envIDStateAndPlan.State == nil {
		// previously was not in an environment, adding
		err := Update(func() (*http.Response, error) {
			return apiClient.AddResourcesToEnvironment(ctx, *plan, client.AddResourcesToEnvironmentJSONRequestBody{
				ResourceIds: []string{resourceID},
			})
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("could not add service to new environment: %w", err)
		}

		return plan, nil
	} else {
		// moving from one environment to another
		err := Update(func() (*http.Response, error) {
			return apiClient.RemoveResourcesFromEnvironment(
				ctx, resourceID, &client.RemoveResourcesFromEnvironmentParams{
					ResourceIds: []string{resourceID},
				})
		}, nil)

		if err != nil {
			return nil, fmt.Errorf("could not remove service from environment for move: %w", err)
		}

		var resp client.Environment
		err = Update(func() (*http.Response, error) {
			return apiClient.AddResourcesToEnvironment(ctx, *plan, client.AddResourcesToEnvironmentJSONRequestBody{
				ResourceIds: []string{resourceID},
			})
		}, &resp)
		if err != nil {
			return nil, fmt.Errorf("could not add service to environment for move: %w", err)
		}

		return &resp.Id, nil
	}
}

func updateCustomDomains(
	ctx context.Context, apiClient *client.ClientWithResponses, serviceID string, cdsStateAndPlan CustomDomainStateAndPlan,
) error {
	cdsToAdd, _, cdsToRemove := XORStringSlices(
		namesForCustomDomainModels(cdsStateAndPlan.Plan),
		namesForCustomDomainModels(cdsStateAndPlan.State),
	)

	for _, name := range cdsToAdd {
		err := Create(func() (*http.Response, error) {
			return apiClient.CreateCustomDomain(ctx, serviceID, client.CreateCustomDomainJSONRequestBody{
				Name: name,
			})
		}, nil)
		if err != nil {
			return fmt.Errorf("could not add custom domain: %w", err)
		}
	}

	for _, name := range cdsToRemove {
		err := Delete(func() (*http.Response, error) {
			return apiClient.DeleteCustomDomain(ctx, serviceID, name)
		})
		if err != nil {
			return fmt.Errorf("could not remove custom domain: %w", err)
		}
	}

	return nil
}

func namesForCustomDomainModels(cd []CustomDomainModel) []string {
	var names []string
	for _, c := range cd {
		names = append(names, c.Name.ValueString())
	}
	return names
}

package common

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/notifications"
)

type WrappedStaticSite struct {
	*client.Service
	CustomDomains        *[]client.CustomDomain
	EnvVars              *[]client.EnvVarWithCursor
	Headers              *[]client.Header
	NotificationOverride *notifications.NotificationOverride
	Routes               *[]client.Route
}

type UpdateStaticSiteReq struct {
	ServiceID            string
	Service              client.UpdateServiceJSONRequestBody
	CustomDomains        CustomDomainStateAndPlan
	EnvironmentID        *EnvironmentIDStateAndPlan
	EnvVars              client.EnvVarInputArray
	Headers              []client.HeaderInput
	NotificationOverride *notifications.NotificationServiceOverridePATCH
	Routes               []client.RoutePut
}

func WrapStaticSite(ctx context.Context, apiClient *client.ClientWithResponses, service *client.Service) (*WrappedStaticSite, error) {
	wrappedService, err := WrapService(ctx, apiClient, service)
	if err != nil {
		return nil, err
	}

	routes, err := getRoutes(ctx, apiClient, service.Id)
	if err != nil {
		return nil, err
	}

	headers, err := getHeaders(ctx, apiClient, service.Id)
	if err != nil {
		return nil, err
	}

	return &WrappedStaticSite{
		Service:              service,
		CustomDomains:        wrappedService.CustomDomains,
		EnvVars:              wrappedService.EnvVars,
		Headers:              headers,
		NotificationOverride: wrappedService.NotificationOverride,
		Routes:               routes,
	}, nil

}

func UpdateStaticSite(ctx context.Context, apiClient *client.ClientWithResponses, skipDeploy bool, req UpdateStaticSiteReq) (*WrappedStaticSite, error) {
	wrappedService, err := UpdateService(ctx, apiClient, skipDeploy, UpdateServiceReq{
		ServiceID:            req.ServiceID,
		Service:              req.Service,
		CustomDomains:        req.CustomDomains,
		EnvVars:              req.EnvVars,
		EnvironmentID:        req.EnvironmentID,
		NotificationOverride: req.NotificationOverride,
	}, ServiceTypeStaticSite)
	if err != nil {
		return nil, err
	}

	var headers []client.Header
	if req.Headers != nil {
		headers, err = updateHeaders(ctx, apiClient, req)
		if err != nil {
			return nil, err
		}
	}

	var routes []client.Route
	if req.Routes != nil {
		routes, err = updateRoutes(ctx, apiClient, req)
		if err != nil {
			return nil, err
		}
	}

	return &WrappedStaticSite{
		Service:              wrappedService.Service,
		CustomDomains:        wrappedService.CustomDomains,
		EnvVars:              wrappedService.EnvVars,
		Headers:              &headers,
		NotificationOverride: wrappedService.NotificationOverride,
		Routes:               &routes,
	}, nil
}

func getHeaders(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*[]client.Header, error) {
	var limit = 100
	var headers []client.Header
	var cursor *string

	for {
		c := cursor

		var headerResp []client.HeaderWithCursor
		err := Get(func() (*http.Response, error) {
			return apiClient.ListHeaders(ctx, serviceID, &client.ListHeadersParams{
				Cursor: c,
				Limit:  &limit,
			})
		}, &headerResp)

		if err != nil {
			return nil, err
		}

		headers = append(headers, HeaderResponseToClientHeaders(headerResp)...)
		if len(headerResp) < limit {
			break
		}

		newCursor := headerResp[len(headerResp)-1].Cursor
		cursor = &newCursor
	}

	return &headers, nil
}

func getRoutes(ctx context.Context, apiClient *client.ClientWithResponses, serviceID string) (*[]client.Route, error) {
	var limit = 100
	var routes []client.Route
	var cursor string

	for {
		c := cursor

		var routeResp []client.RouteWithCursor
		err := Get(func() (*http.Response, error) {
			return apiClient.ListRoutes(ctx, serviceID, &client.ListRoutesParams{
				Cursor: &c,
				Limit:  &limit,
			})
		}, &routeResp)
		if err != nil {
			return nil, err
		}

		routes = append(routes, RouteResponseToClientRoutes(routeResp)...)

		if len(routeResp) < limit {
			break
		}

		newCursor := routeResp[len(routeResp)-1].Cursor
		cursor = newCursor
	}

	return &routes, nil
}

func updateHeaders(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateStaticSiteReq) ([]client.Header, error) {
	var headerResp []client.Header
	err := Update(func() (*http.Response, error) {
		return apiClient.UpdateHeaders(ctx, req.ServiceID, req.Headers)
	}, &headerResp)
	if err != nil {
		return nil, fmt.Errorf("could not update headers: %w", err)
	}

	return headerResp, nil
}

func updateRoutes(ctx context.Context, apiClient *client.ClientWithResponses, req UpdateStaticSiteReq) ([]client.Route, error) {
	var routeResp []client.Route
	err := Update(func() (*http.Response, error) {
		return apiClient.PutRoutes(ctx, req.ServiceID, req.Routes)
	}, &routeResp)
	if err != nil {
		return nil, fmt.Errorf("could not update routes: %w", err)
	}

	return routeResp, nil
}

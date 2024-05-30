package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

type RouteModel struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Type        types.String `tfsdk:"type"`
}

func RouteResponseToClientRoutes(routeResp []client.RouteWithCursor) []client.Route {
	routes := make([]client.Route, len(routeResp))
	for i, rr := range routeResp {
		routes[i] = client.Route{
			Id:          rr.Route.Id,
			Source:      rr.Route.Source,
			Destination: rr.Route.Destination,
			Type:        rr.Route.Type,
		}
	}
	return routes
}

func RouteModelToClientRoutePutInput(routes []RouteModel) []client.RoutePut {
	clientRoutes := make([]client.RoutePut, len(routes))
	for i, r := range routes {
		clientRoutes[i] = client.RoutePut{
			Source:      r.Source.ValueString(),
			Destination: r.Destination.ValueString(),
			Type:        ClientRouteType(r.Type.ValueString()),
		}
	}
	return clientRoutes
}

func ClientRoutesToRouteModels(route []client.Route) []RouteModel {
	routes := make([]RouteModel, len(route))
	for i, r := range route {
		routes[i] = RouteModel{
			Source:      types.StringValue(r.Source),
			Destination: types.StringValue(r.Destination),
			Type:        types.StringValue(string(r.Type)),
		}
	}

	return routes
}

func ClientRouteType(routeType string) client.RouteType {
	switch routeType {
	case "redirect":
		return client.RouteTypeRedirect
	case "rewrite":
		return client.RouteTypeRewrite
	}
	return ""
}

func SortRoutesForPlan(planRoutes []RouteModel, staticSiteRoutes []RouteModel) ([]RouteModel, error) {
	// planRoutes will be nil when importing an existing resource. In this case, the sort order does not matter.
	if planRoutes == nil {
		return staticSiteRoutes, nil
	}

	var routes []RouteModel
	for _, planRoute := range planRoutes {
		for _, route := range staticSiteRoutes {
			if planRoute.Source.String() == route.Source.String() {
				routes = append(routes, route)
				break
			}
		}
	}

	// should not happen
	if len(routes) != len(planRoutes) {
		return nil, fmt.Errorf("failed to sort routes for plan")
	}

	return routes, nil
}

package project

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
)

type ProjectModel struct {
	Id           types.String                 `tfsdk:"id"`
	Name         types.String                 `tfsdk:"name"`
	Environments map[string]*EnvironmentModel `tfsdk:"environments"`
}

type EnvironmentModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ProtectedStatus types.String `tfsdk:"protected_status"`
	NetworkIsolated types.Bool   `tfsdk:"network_isolated"`
	IPAllowList     types.Set    `tfsdk:"ip_allow_list"`
}

func ClientProtectedStatusFromModel(env EnvironmentModel) client.ProtectedStatus {
	protectedStatus := client.Unprotected
	if env.ProtectedStatus.ValueString() == string(client.Protected) {
		protectedStatus = client.Protected
	}
	return protectedStatus
}

func ModelForEnvironmentResult(env *client.Environment, plan *EnvironmentModel, diags diag.Diagnostics) EnvironmentModel {
	// Handle IP allow list: if not configured in plan (null), keep it null
	// This prevents showing drift when API returns its default value
	var ipAllowList types.Set
	if plan != nil && !plan.IPAllowList.IsNull() {
		if env.IpAllowList != nil {
			ipAllowList = common.IPAllowListFromClient(*env.IpAllowList, diags)
		} else {
			// API returned null but we have it in plan, keep the plan value
			ipAllowList = plan.IPAllowList
		}
	} else {
		// Not configured in plan, keep it null
		ipAllowList = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cidr_block":  types.StringType,
				"description": types.StringType,
			},
		})
	}

	return EnvironmentModel{
		Id:              types.StringValue(env.Id),
		Name:            types.StringValue(env.Name),
		ProtectedStatus: types.StringValue(string(env.ProtectedStatus)),
		NetworkIsolated: types.BoolValue(env.NetworkIsolationEnabled),
		IPAllowList:     ipAllowList,
	}
}

func ModelForProjectResult(project *client.Project, environments map[string]*client.Environment, planEnvironments map[string]*EnvironmentModel, diags diag.Diagnostics) (ProjectModel, error) {
	environmentsList := make(map[string]*EnvironmentModel)
	for k, env := range environments {
		// Get the plan for this environment if it exists
		var planEnv *EnvironmentModel
		if planEnvironments != nil {
			planEnv = planEnvironments[k]
		}
		resEnv := common.From(ModelForEnvironmentResult(env, planEnv, diags))
		environmentsList[k] = resEnv
	}

	projectModel := ProjectModel{
		Id:           types.StringValue(project.Id),
		Name:         types.StringValue(project.Name),
		Environments: environmentsList,
	}

	return projectModel, nil
}

func Read(ctx context.Context, c *client.ClientWithResponses, proj ProjectModel) (*ProjectModel, error) {
	var projectResponse client.Project
	err := common.Get(func() (*http.Response, error) {
		return c.RetrieveProject(ctx, proj.Id.ValueString())
	}, &projectResponse)

	if err != nil {
		return nil, err
	}

	environments := make(map[string]*client.Environment)

	for _, envId := range projectResponse.EnvironmentIds {
		var environmentResponse *client.Environment

		err := common.Get(func() (*http.Response, error) {
			return c.RetrieveEnvironment(ctx, envId)
		}, &environmentResponse)

		if err != nil {
			return nil, err
		}

		var key string
		for k, env := range proj.Environments {
			if env.Id.ValueString() == envId {
				key = k
				break
			}
		}

		// If there is no key in the state for this environment, we are likely trying to
		// import the resource. In this case, we will use the environment name as the key.
		if len(key) == 0 {
			key = environmentResponse.Name
		}

		environments[key] = environmentResponse
	}

	var diags diag.Diagnostics
	projectModel, err := ModelForProjectResult(&projectResponse, environments, proj.Environments, diags)
	if err != nil {
		return nil, err
	}
	return &projectModel, nil
}

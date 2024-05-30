package checks

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var _ plancheck.PlanCheck = expectNoReplace{}

type expectNoReplace struct{}

func (e expectNoReplace) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	var result error
	for _, rc := range req.Plan.ResourceChanges {
		if rc.Change.Actions.DestroyBeforeCreate() {
			result = errors.Join(result, fmt.Errorf("expected no replacement, but %s has planned destroy-before-create", rc.Address))
		}
	}

	resp.Error = result
}

func ExpectNoReplace() plancheck.PlanCheck {
	return expectNoReplace{}
}

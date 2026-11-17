package resource_test

import (
	"context"
	"testing"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/postgres"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/redis"
	tfresource "terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestPlanValidatorsAcceptGeneratedPlans drives every plan name the generated
// client knows about through the validator the corresponding resource attribute
// uses. Sourcing the cases from the generated *Values() accessors means new plan
// names added to the OpenAPI schema are covered by regeneration alone.
func TestPlanValidatorsAcceptGeneratedPlans(t *testing.T) {
	for _, tc := range []struct {
		name      string
		validator validator.String
		plans     []string
	}{
		{
			name:      "postgres",
			validator: tfresource.ValidatePostgresPlanFunc(),
			// "custom" is excluded throughout: a custom plan is named by its
			// own identifier, covered by TestPlanValidatorsRejectUnknownPlans.
			plans: common.EnumStringsExcept(postgres.PostgresPlansValues(), postgres.Custom),
		},
		{
			// The Key Value attribute validates with the Redis plan validator,
			// so drive the Key Value values through it: divergence between the
			// two enums fails here.
			name:      "key value",
			validator: redis.ValidateRedisPlanFunc(),
			plans:     common.EnumStringsExcept(client.KeyValuePlanValues(), client.KeyValuePlanCustom),
		},
		{
			name:      "redis",
			validator: redis.ValidateRedisPlanFunc(),
			plans:     common.EnumStringsExcept(client.RedisPlanValues(), client.RedisPlanCustom),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotEmpty(t, tc.plans, "no generated plan values to validate")

			for _, plan := range tc.plans {
				resp := &validator.StringResponse{}
				tc.validator.ValidateString(
					context.Background(),
					validator.StringRequest{ConfigValue: types.StringValue(plan)},
					resp,
				)

				assert.False(t, resp.Diagnostics.HasError(), "plan %q was rejected", plan)
			}
		})
	}
}

func TestPlanValidatorsRejectUnknownPlans(t *testing.T) {
	for _, tc := range []struct {
		name      string
		validator validator.String
		plan      string
	}{
		{name: "postgres", validator: tfresource.ValidatePostgresPlanFunc(), plan: "not-a-plan"},
		{name: "redis", validator: redis.ValidateRedisPlanFunc(), plan: "not-a-plan"},
		// A custom plan is named by its own identifier, so the bare enum member
		// is not something a config can ask for.
		{
			name:      "postgres custom",
			validator: tfresource.ValidatePostgresPlanFunc(),
			plan:      string(postgres.Custom),
		},
		{
			name:      "redis custom",
			validator: redis.ValidateRedisPlanFunc(),
			plan:      string(client.RedisPlanCustom),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			tc.validator.ValidateString(
				context.Background(),
				validator.StringRequest{ConfigValue: types.StringValue(tc.plan)},
				resp,
			)

			assert.True(t, resp.Diagnostics.HasError(), "expected plan %q to be rejected", tc.plan)
		})
	}
}

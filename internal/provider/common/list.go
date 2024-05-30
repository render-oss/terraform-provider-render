package common

import "github.com/hashicorp/terraform-plugin-framework/types"

func ConvertTypesSetToStringSlice(tSet *types.Set) *[]string {
	if tSet.IsNull() {
		return nil
	}

	result := make([]string, 0, len(tSet.Elements()))

	for _, elem := range tSet.Elements() {
		if strValue, ok := elem.(types.String); ok && !strValue.IsNull() {
			result = append(result, strValue.String())
		}
	}

	return &result
}

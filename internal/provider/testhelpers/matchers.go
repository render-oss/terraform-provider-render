package testhelpers

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func CheckIDPrefix(idPrefix string) resource.CheckResourceAttrWithFunc {
	return func(value string) error {
		if !strings.HasPrefix(value, idPrefix) {
			return fmt.Errorf("expected id to have %s prefix, got: %s", idPrefix, value)
		}
		return nil
	}
}

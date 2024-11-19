package datasource

import (
	"terraform-provider-render/internal/provider/postgres"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var DiskSizeGB schema.Int64Attribute = schema.Int64Attribute{
	Description: "Disk size in GB.",
	Computed:    true,
	Optional:    true,
	Validators:  []validator.Int64{postgres.ValidateDiskSizeGB()},
}

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

func DiskToClientPatch(disk DiskModel) client.DiskPATCH {
	return client.DiskPATCH{
		SizeGB:    CastPointerToInt(disk.SizeGB.ValueInt64Pointer()),
		MountPath: disk.MountPath.ValueStringPointer(),
		Name:      disk.Name.ValueStringPointer(),
	}
}

func DiskToClientPOST(serviceID string, disk DiskModel) client.DiskPOST {
	return client.DiskPOST{
		ServiceId: serviceID,
		SizeGB:    int(disk.SizeGB.ValueInt64()),
		MountPath: disk.MountPath.ValueString(),
		Name:      disk.Name.ValueString(),
	}
}

func DiskToClientCreate(disk *DiskModel) *client.ServiceDisk {
	if disk == nil {
		return nil
	}

	return &client.ServiceDisk{
		SizeGB:    CastPointerToInt(disk.SizeGB.ValueInt64Pointer()),
		MountPath: disk.MountPath.ValueString(),
		Name:      disk.Name.ValueString(),
	}
}

func DiskDetailsToDiskModel(disk *client.DiskDetails) *DiskModel {
	if disk == nil {
		return nil
	}

	return &DiskModel{
		ID:        types.StringValue(disk.Id),
		Name:      types.StringValue(disk.Name),
		SizeGB:    types.Int64Value(int64(disk.SizeGB)),
		MountPath: types.StringValue(disk.MountPath),
	}
}

func DiskToDiskModel(disk *client.Disk) *DiskModel {
	if disk == nil {
		return nil
	}

	return &DiskModel{
		ID:        types.StringValue(disk.Id),
		Name:      types.StringValue(disk.Name),
		SizeGB:    types.Int64Value(int64(disk.SizeGB)),
		MountPath: types.StringValue(disk.MountPath),
	}
}

func DiskDetailsToDisk(disk *client.DiskDetails) *client.Disk {
	if disk == nil {
		return nil
	}

	return &client.Disk{
		Id:        disk.Id,
		Name:      disk.Name,
		SizeGB:    disk.SizeGB,
		MountPath: disk.MountPath,
	}
}

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/client/disks"
)

func DiskToClientPatch(disk DiskModel) disks.DiskPATCH {
	return disks.DiskPATCH{
		SizeGB:    CastPointerToInt(disk.SizeGB.ValueInt64Pointer()),
		MountPath: disk.MountPath.ValueStringPointer(),
		Name:      disk.Name.ValueStringPointer(),
	}
}

func DiskToClientPOST(serviceID string, disk DiskModel) disks.DiskPOST {
	return disks.DiskPOST{
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

func DiskDetailsToDiskModel(disk *disks.DiskDetails) *DiskModel {
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

func DiskToDiskModel(disk *disks.Disk) *DiskModel {
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

func DiskDetailsToDisk(disk *disks.DiskDetails) *disks.Disk {
	if disk == nil {
		return nil
	}

	return &disks.Disk{
		Id:        disk.Id,
		Name:      disk.Name,
		SizeGB:    disk.SizeGB,
		MountPath: disk.MountPath,
	}
}

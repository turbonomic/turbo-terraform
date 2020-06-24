package dtos

import (
	"github.com/enlinxu/turbo-terraform/pkg/discovery/constant"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder/group"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

func CreateVMEntityDto(name string, id string, entityPropertyName string, workloadControllerId string) (*proto.EntityDTO, error) {
	entityDto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_VIRTUAL_MACHINE, id).
		DisplayName(name).
		ControlledBy(workloadControllerId).
		WithProperty(getEntityProperty(entityPropertyName)).
		ReplacedBy(getReplacementMetaData(proto.EntityDTO_VIRTUAL_MACHINE)).
		Monitored(false).
		Create()
	if err != nil {
		glog.Errorf("Error building EntityDTO for name %s: %s", name, err)
		return nil, err
	}
	return entityDto, nil
}

func CreateGroupDto(path string, name string, instances []string) (*proto.GroupDTO, error) {
	groupBuilder := group.StaticGroup(path + name).
		OfType(proto.EntityDTO_VIRTUAL_MACHINE).
		WithEntities(instances).
		WithDisplayName(path + name).ResizeConsistently()
	groupDTO, err := groupBuilder.Build()
	if err != nil {
		glog.Errorf("Error creating group dto  %s::%s", name, err)
		return nil, err
	}

	return groupDTO, nil
}

func CreateWorkloadControllerDto(name string) (*proto.EntityDTO, error) {
	entityDto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_WORKLOAD_CONTROLLER, name).
		DisplayName(name).
		Monitored(true).
		Create()
	if err != nil {
		glog.Errorf("Error building EntityDTO for name %s: %s", name, err)
		return nil, err
	}
	return entityDto, nil
}

func getReplacementMetaData(entityType proto.EntityDTO_EntityType) *proto.EntityDTO_ReplacementEntityMetaData {
	attr := constant.SUPPLY_CHAIN_CONSTANT_ID
	useTopoExt := true

	b := builder.NewReplacementEntityMetaDataBuilder().
		Matching("Proxy_VM_UUID").
		MatchingExternal(&proto.ServerEntityPropDef{
			Entity:     &entityType,
			Attribute:  &attr,
			UseTopoExt: &useTopoExt,
		})

	return b.Build()
}

func getEntityProperty(value string) *proto.EntityDTO_EntityProperty {
	attr := constant.STITCHING_PROPERTY_NAME
	ns := constant.DefaultPropertyNamespace

	return &proto.EntityDTO_EntityProperty{
		Namespace: &ns,
		Name:      &attr,
		Value:     &value,
	}
}

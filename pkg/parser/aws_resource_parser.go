package parser

import (
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/discovery/dtos"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type AwsInstance struct {
	instanceId        string
	availability_zone string
}

type AwsInstanceResource struct {
	id        string
	instances []AwsInstance
}

type AwsParser struct {
	resource             *Resource
	tfStatePath          string
	workloadControllerId string
	assets               map[string]struct{}
}

func NewAwsParser(resource *Resource, path string, workloadControllerId string, assets map[string]struct{}) *AwsParser {
	return &AwsParser{
		resource:             resource,
		tfStatePath:          path,
		workloadControllerId: workloadControllerId,
		assets:               assets,
	}
}

func (parser *AwsParser) ParseAwsInstanceResource(entityToAssetsMap map[string]map[string]struct{}) ([]*proto.EntityDTO, []*proto.GroupDTO, error) {
	var entityDTOs []*proto.EntityDTO
	var groupDTOs []*proto.GroupDTO
	name := parser.resource.Name
	members := []string{}
	workloadControllerId := parser.workloadControllerId
	for _, instance := range parser.resource.Instances {
		attributes := instance.Attributes
		id := fmt.Sprintf("%v", attributes["id"])
		availabilityZone := fmt.Sprintf("%v", attributes["availability_zone"])
		entityPropertyName := getAwsInstanceName(id, availabilityZone)
		entityDto, e := dtos.CreateEntityDto(proto.EntityDTO_VIRTUAL_MACHINE, name, id, entityPropertyName, workloadControllerId)
		if e != nil {
			glog.Errorf("Error building EntityDTO from metric %s", e)
			return nil, nil, e
		}
		entityDTOs = append(entityDTOs, entityDto)
		entityToAssetsMap[entityPropertyName] = parser.assets
		members = append(members, id)
	}

	if len(parser.resource.Instances) > 1 {
		//For the group name here, use the directory of the tf state location.
		groupDto, e := dtos.CreateGroupDto(parser.tfStatePath, name, members)
		if e != nil {
			glog.Errorf("Error building groupDTO from metric %s", e)
			return nil, nil, e
		}
		groupDTOs = append(groupDTOs, groupDto)
	}

	return entityDTOs, groupDTOs, nil
}

func (parser *AwsParser) ParseAwsASGResource() ([]*proto.EntityDTO, error) {
	var entityDTOs []*proto.EntityDTO
	name := parser.resource.Name
	workloadControllerId := parser.workloadControllerId
	for _, instance := range parser.resource.Instances {
		attributes := instance.Attributes
		id := fmt.Sprintf("%v", attributes["id"])
		proxy_id := fmt.Sprintf("%v", attributes["arn"])
		entityDto, e := dtos.CreateEntityDto(proto.EntityDTO_VM_SPEC, name, id, proxy_id, workloadControllerId)
		if e != nil {
			glog.Errorf("Error building EntityDTO from metric %s", e)
			return nil, e
		}
		entityDTOs = append(entityDTOs, entityDto)
	}
	return entityDTOs, nil
}

func getAwsInstanceName(id string, az string) string {
	awsFormat := "aws::%v::VM::%v"
	region := az[0 : len(az)-1]
	result := fmt.Sprintf(awsFormat, region, id)
	return result
}

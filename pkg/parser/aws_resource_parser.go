package parser

import (
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/discovery/dtos"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
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
	files                map[string]struct{}
}

func NewAwsParser(resource *Resource, path string, workloadControllerId string, files map[string]struct{}) *AwsParser {
	return &AwsParser{
		resource:             resource,
		tfStatePath:          path,
		workloadControllerId: workloadControllerId,
		files:                files,
	}
}

func (parser *AwsParser) GetAwsInstanceResource(entityToFilesMap map[string]map[string]struct{}) ([]*proto.EntityDTO, []*proto.GroupDTO, []string, error) {
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
		entityDto, e := dtos.CreateVMEntityDto(name, id, entityPropertyName, workloadControllerId)
		if e != nil {
			glog.Errorf("Error building EntityDTO from metric %s", e)
			return nil, nil, nil, e
		}
		entityDTOs = append(entityDTOs, entityDto)
		entityToFilesMap[entityPropertyName] = parser.files
		members = append(members, id)
	}

	if len(parser.resource.Instances) > 1 {
		//For the group name here, use the directory of the tf state location.
		groupDto, e := dtos.CreateGroupDto(parser.tfStatePath[:strings.LastIndex(parser.tfStatePath, "/")+1], name, members)
		if e != nil {
			glog.Errorf("Error building groupDTO from metric %s", e)
			return nil, nil, nil, e
		}
		groupDTOs = append(groupDTOs, groupDto)
	}

	return entityDTOs, groupDTOs, members, nil
}

func getAwsInstanceName(id string, az string) string {
	awsFormat := "aws::%v::VM::%v"
	region := az[0 : len(az)-1]
	result := fmt.Sprintf(awsFormat, region, id)
	return result
}

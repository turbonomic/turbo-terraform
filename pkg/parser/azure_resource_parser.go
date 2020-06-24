package parser

import (
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/discovery/dtos"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
)

type AzureInstance struct {
	instanceId string
}

type AzureInstanceResource struct {
	id        string
	instances []AwsInstance
}

type AzureParser struct {
	resource             *Resource
	tfStatePath          string
	workloadControllerId string
	files                map[string]struct{}
}

func NewAzureParser(resource *Resource, path string, workloadControllerId string, files map[string]struct{}) *AzureParser {
	return &AzureParser{
		resource:             resource,
		tfStatePath:          path,
		workloadControllerId: workloadControllerId,
		files:                files,
	}
}

func (parser *AzureParser) GetAzureInstanceResource(entityToFilesMap map[string]map[string]struct{}) ([]*proto.EntityDTO, []*proto.GroupDTO, error) {
	var entityDTOs []*proto.EntityDTO
	var groupDTOs []*proto.GroupDTO
	name := parser.resource.Name
	members := []string{}

	for _, instance := range parser.resource.Instances {
		attributes := instance.Attributes
		id := fmt.Sprintf("%v", attributes["virtual_machine_id"])
		entityPropertyName := getAzureInstanceName(id)
		entityDto, e := dtos.CreateVMEntityDto(name, id, entityPropertyName, parser.workloadControllerId)
		if e != nil {
			glog.Errorf("Error building EntityDTO from metric %s", e)
			return nil, nil, e
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
			return nil, nil, e
		}
		groupDTOs = append(groupDTOs, groupDto)
	}

	return entityDTOs, groupDTOs, nil
}

func getAzureInstanceName(id string) string {
	azureFormat := "azure::VM::%v"
	result := fmt.Sprintf(azureFormat, id)
	return result
}

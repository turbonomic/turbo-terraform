package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/discovery/dtos"
	"github.com/enlinxu/turbo-terraform/pkg/parser"
	"github.com/enlinxu/turbo-terraform/pkg/query"
	"github.com/enlinxu/turbo-terraform/pkg/registration"
	"github.com/enlinxu/turbo-terraform/pkg/util"
	"github.com/golang/glog"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"strings"
)

//TODO: Make the map threadsafe, maybe looking at the syncmap
var EntityIdToAssetsMap = make(map[string]map[string]struct{})

type DiscoveryClient struct {
	targetConfig *TargetConf
	tfPath       *string
	tfToken      *string
	org          *string
}

type DiscoveryTargetParams struct {
	OptionalTargetAddress *string
	TargetType            string
	TargetName            string
	ProbeCategory         string
}

type EntityBuilderParams struct {
	keepStandalone bool
}

type AwsInstance struct {
	id                string
	availability_zone string
}

type CurrentStateVersionResult struct {
	Data *DataInfo `json:"data,omitempty"`
}
type DataInfo struct {
	Id         string                 `json:"id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func NewDiscoveryClient(targetConfig *TargetConf, tfPath *string, tfToken *string, org *string) *DiscoveryClient {
	return &DiscoveryClient{
		targetConfig: targetConfig,
		tfPath:       tfPath,
		tfToken:      tfToken,
		org:          org,
	}
}

func (dc *DiscoveryClient) String() string {
	return fmt.Sprintf("%+v", dc.targetConfig)
}

func (dc *DiscoveryClient) GetAccountValues() *sdkprobe.TurboTargetInfo {
	var accountValues []*proto.AccountValue

	targetConf := dc.targetConfig
	// Convert all parameters in clientConf to AccountValue list
	targetID := registration.TargetIdField
	accVal := &proto.AccountValue{
		Key:         &targetID,
		StringValue: &targetConf.Identifier,
	}
	accountValues = append(accountValues, accVal)

	targetInfo := sdkprobe.NewTurboTargetInfoBuilder(targetConf.ProbeCategory, targetConf.TargetType, targetID, accountValues).Create()

	glog.V(2).Infof("Got AccountValues for target:%v", targetConf.Identifier)
	return targetInfo
}

func (dc *DiscoveryClient) Validate(accountValues []*proto.AccountValue) (*proto.ValidationResponse, error) {
	glog.V(2).Infof("begin to validating target...")
	return &proto.ValidationResponse{}, nil
}

func printDTOs(dtos []*proto.EntityDTO) string {
	msg := ""
	for _, dto := range dtos {
		line := fmt.Sprintf("%+v", dto)
		msg = msg + "\n" + line
	}

	return msg
}

func (dc *DiscoveryClient) Discover(accountValues []*proto.AccountValue) (*proto.DiscoveryResponse, error) {
	glog.V(2).Infof("begin to discovery target...")

	var entityDTOs []*proto.EntityDTO
	var groupDTOs []*proto.GroupDTO
	if *dc.tfPath != "" {
		tfStateTofiles, err := util.CreateTFStateToFilesMap(*dc.tfPath, "*.tfstate")
		if err != nil {
			glog.Error("Failed to parse the TF State files %v" + err.Error())
			return nil, err
		}
		for tfStateFilePath, files := range tfStateTofiles {
			tfstate, e := parser.ParseTerraformStateFile(tfStateFilePath)
			if e != nil {
				return nil, fmt.Errorf("File error: %v\n" + e.Error())
			}
			resources := tfstate.Resources
			dirPath := tfStateFilePath[:strings.LastIndex(tfStateFilePath, "/")+1]
			//Create one workload controller per TFState
			wcDto, e := dtos.CreateWorkloadControllerDto(dirPath, dirPath)
			//wcDto, e := dtos.CreateVmSpecDto(dirPath)
			if e != nil {
				glog.Errorf("Error building workload controller from metric %s", e)
				return nil, err
			}
			entityDTOS, groupDTOS, e := parseResources(resources, tfStateFilePath[:strings.LastIndex(tfStateFilePath, "/")+1], dirPath, files)
			entityDTOs = append(entityDTOs, wcDto)
			entityDTOs = append(entityDTOs, entityDTOS...)
			groupDTOs = append(groupDTOs, groupDTOS...)
		}

	}
	if *dc.tfToken != "" {
		//When token is not null, meaning it's terraform enterprise
		//Step 1 Get the Workspaces
		result, err := query.GetWorkspaces(*dc.org, *dc.tfToken)
		if err != nil {
			glog.Errorf("Error Getting workspace list %s", err)
			return nil, err
		}
		var workspaceResults query.QueryResultInfoArray
		err = json.Unmarshal(result, &workspaceResults)
		if err != nil {
			glog.Errorf("Error unmarshalling Workspaces", err)
		}
		for _, result := range workspaceResults.Data {
			//Step 2 Get current state
			id := result.Id
			workspaceName := fmt.Sprintf("%v", result.Attributes["name"])
			//Step 3 Get all the variables in the workspace
			r, err := query.GetVariables(*dc.org, workspaceName, *dc.tfToken)
			if err != nil {
				glog.Errorf("Error Getting variables", err)
			}
			var variablesResults query.QueryResultInfoArray
			err = json.Unmarshal(r, &variablesResults)
			if err != nil {
				glog.Errorf("Error unmarshalling variables", err)
			}
			variableIds := make(map[string]struct{})

			for _, variable := range variablesResults.Data {
				variableIds[variable.Id] = struct{}{}
			}
			stateResult, err := query.GetCurrentStateVersion(id, *dc.tfToken)
			if err != nil {
				glog.Errorf("Error Getting current state", err)
			}
			tfstate, e := parser.ParseTerraformStateAPI(stateResult)
			if e != nil {
				return nil, fmt.Errorf("File error: %v\n" + e.Error())
			}
			resources := tfstate.Resources
			//Create one workload controller per TFState
			wcDto, e := dtos.CreateWorkloadControllerDto(id, workspaceName)
			//wcDto, e := dtos.CreateVmSpecDto(dirPath)
			if e != nil {
				glog.Errorf("Error building workload controller from metric %s", e)
				return nil, err
			}
			entityDTOS, groupDTOS, e := parseResources(resources, workspaceName, id, variableIds)
			entityDTOs = append(entityDTOs, wcDto)
			entityDTOs = append(entityDTOs, entityDTOS...)
			groupDTOs = append(groupDTOs, groupDTOS...)
		}
	}

	glog.V(2).Infof("end of discoverying target. [%d]", len(entityDTOs))
	glog.V(4).Infof("DTOs:\n%s", printDTOs(entityDTOs))
	response := &proto.DiscoveryResponse{
		EntityDTO:       entityDTOs,
		DiscoveredGroup: groupDTOs,
	}
	return response, nil
}

func parseResources(resources []*parser.Resource, workspaceName string, id string, files map[string]struct{}) ([]*proto.EntityDTO, []*proto.GroupDTO, error) {
	var entityDTOs []*proto.EntityDTO
	var groupDTOs []*proto.GroupDTO
	for _, resource := range resources {
		if strings.HasPrefix(resource.Type, "aws") {
			awsParser := parser.NewAwsParser(resource, workspaceName, id, files)
			if resource.Type == "aws_instance" {
				awsEntityDtos, awsGroupDTOS, e := awsParser.ParseAwsInstanceResource(EntityIdToAssetsMap)
				if e != nil {
					glog.Errorf("Error building EntityDTO and GroupDTO for AWS Instances %s", e)
					return nil, nil, e
				}
				entityDTOs = append(entityDTOs, awsEntityDtos...)
				groupDTOs = append(groupDTOs, awsGroupDTOS...)
			} else if resource.Type == "aws_autoscaling_group" {
				awsEntityDtos, e := awsParser.ParseAwsASGResource()
				if e != nil {
					glog.Errorf("Error building EntityDTO for AWS ASG %s", e)
					return nil, nil, e
				}
				entityDTOs = append(entityDTOs, awsEntityDtos...)
			}
		} else if resource.Type == "azurerm_linux_virtual_machine" || resource.Type == "azurerm_windows_virtual_machine" {
			azureParser := parser.NewAzureParser(resource, workspaceName, id, files)
			azureEntityDtos, azureGroupDTOS, e := azureParser.GetAzureInstanceResource(EntityIdToAssetsMap)
			if e != nil {
				glog.Errorf("Error building EntityDTO and GroupDTO for AZURE Instances %s", e)
				return nil, nil, e
			}
			entityDTOs = append(entityDTOs, azureEntityDtos...)
			groupDTOs = append(groupDTOs, azureGroupDTOS...)
		} else if resource.Type == "vsphere_virtual_machine" {
			if resource.Name == "template" {
				continue
			}
			vsParser := parser.NewVSphereParser(resource, workspaceName, id, files)
			vmEntityDtos, vmGroupDTOS, e := vsParser.ParseVSphereInstanceResource(EntityIdToAssetsMap)
			if e != nil {
				glog.Errorf("Error building EntityDTO and GroupDTO for AZURE Instances %s", e)
				return nil, nil, e
			}
			entityDTOs = append(entityDTOs, vmEntityDtos...)
			groupDTOs = append(groupDTOs, vmGroupDTOS...)
		}
	}
	return entityDTOs, groupDTOs, nil
}

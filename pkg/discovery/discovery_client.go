package discovery

import (
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/parser"
	"github.com/enlinxu/turbo-terraform/pkg/registration"
	"github.com/enlinxu/turbo-terraform/pkg/util"
	"github.com/golang/glog"

	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

var EntityIdToFilesMap = make(map[string]map[string]struct{})

type DiscoveryClient struct {
	targetConfig *TargetConf
	tfPath       *string
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

func NewDiscoveryClient(targetConfig *TargetConf, tfPath *string) *DiscoveryClient {
	return &DiscoveryClient{
		targetConfig: targetConfig,
		tfPath:       tfPath,
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
	// Replace with real discovery
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

		for _, resource := range resources {
			if resource.Type == "aws_instance" {
				awsParser := parser.NewAwsParser(resource, tfStateFilePath, files)
				awsEntityDtos, awsGroupDTOS, e := awsParser.GetAwsInstanceResource(EntityIdToFilesMap)
				if e != nil {
					glog.Errorf("Error building EntityDTO and GroupDTO for AWS Instances %s", err)
					return nil, err
				}
				entityDTOs = append(entityDTOs, awsEntityDtos...)
				groupDTOs = append(groupDTOs, awsGroupDTOS...)
			}
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

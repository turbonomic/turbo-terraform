package discovery

import (
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/parser"
	"github.com/enlinxu/turbo-terraform/pkg/registration"
	"github.com/enlinxu/turbo-terraform/pkg/util"
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

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

// Implements the TurboDiscoveryClient interface
type TFDiscoveryClient struct {
	discoveryTargetParams *DiscoveryTargetParams
	keepStandalone        bool
	metricEndpoint        []string
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

	var resultDTOs []*proto.EntityDTO

	// Replace with real discovery
	tfStateTofiles, err := util.CreateTFStateToFilesMap(*dc.tfPath, "*.tfstate")
	if err != nil {
		glog.Error("Failed to parse the TF State files %v" + err.Error())
		return nil, err
	}
	for tfStateFilePath, _ := range tfStateTofiles {
		tfstate, e := parser.ParseTerraformStateFile(tfStateFilePath)
		if e != nil {
			return nil, fmt.Errorf("File error: %v\n" + e.Error())
		}
		resources := tfstate.Resources
		var instanceMap map[string]map[AwsInstance]struct{}

		for _, resource := range resources {
			if resource.Type == "aws_instance" {
				instanceMap = createNameToAwsInstancesMap(resource)
			}
		}
		for name, instances := range instanceMap {
			for instance := range instances {
				entityDto, e := createEntityDto(name, instance.id, instance.availability_zone)
				if e != nil {
					glog.Errorf("Error building EntityDTO from metric %s", err)
					return nil, err
				}
				resultDTOs = append(resultDTOs, entityDto)
			}

		}
	}

	glog.V(2).Infof("end of discoverying target. [%d]", len(resultDTOs))
	glog.V(3).Infof("DTOs:\n%s", printDTOs(resultDTOs))
	printDTOs(resultDTOs)
	response := &proto.DiscoveryResponse{
		EntityDTO: resultDTOs,
	}

	return response, nil
}

func createNameToAwsInstancesMap(resource *parser.Resource) map[string]map[AwsInstance]struct{} {
	instanceNameToIntancesMap := make(map[string]map[AwsInstance]struct{})
	name := resource.Name
	instanceSet, exist := instanceNameToIntancesMap[name]
	if !exist {
		instanceSet = make(map[AwsInstance]struct{})
	}
	for _, instance := range resource.Instances {
		attributes := instance.Attributes
		id := attributes["id"]
		availabilityZone := attributes["availability_zone"]
		awsInstance := &AwsInstance{
			id:                fmt.Sprintf("%v", id),
			availability_zone: fmt.Sprintf("%v", availabilityZone),
		}
		instanceSet[*awsInstance] = struct{}{}

	}
	instanceNameToIntancesMap[name] = instanceSet
	return instanceNameToIntancesMap
}

func createEntityDto(name string, id string, az string) (*proto.EntityDTO, error) {
	entityDto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_VIRTUAL_MACHINE, id).
		DisplayName(name).
		WithProperty(getEntityProperty(getAwsInstanceName(id, az))).
		ReplacedBy(getReplacementMetaData(proto.EntityDTO_VIRTUAL_MACHINE)).
		Monitored(false).
		Create()
	if err != nil {
		glog.Errorf("Error building EntityDTO for name %s: %s", name, err)
		return nil, err
	}
	//entityDto.KeepStandalone = true

	return entityDto, nil
}

func getAwsInstanceName(id string, az string) string {
	awsFormat := "aws::%v::VM::%v"
	region := az[0 : len(az)-1]
	result := fmt.Sprintf(awsFormat, region, id)
	return result
}

func getReplacementMetaData(entityType proto.EntityDTO_EntityType) *proto.EntityDTO_ReplacementEntityMetaData {
	attr := "id"
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
	attr := "Proxy_VM_UUID"
	ns := "DEFAULT"

	return &proto.EntityDTO_EntityProperty{
		Namespace: &ns,
		Name:      &attr,
		Value:     &value,
	}
}

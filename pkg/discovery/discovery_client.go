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

type TFState struct {
	Version   string      `json:"version,omitempty"`
	Resources []*Resource `json:"resources,omitempty"`
}

type Resource struct {
	Type      string      `json:"type,omitempty"`
	Name      string      `json:"name,omitempty"`
	Provider  string      `json:"provider,omitempty"`
	Instances []*Instance `json:"instances,omitempty"`
}

type Instance struct {
	IndexKey   string                `json:"index_key,omitempty"`
	Attributes []*InstanceAttributes `json:"attributes,omitempty"`
}

type InstanceAttributes struct {
	Ami              string `json:"ami,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
	CPUCoreCount     int    `json:"cpu_core_count,omitempty"`
	InstanceType     string `json:"instance_type,omitempty"`
	Id               string `json:"id,omitempty"`
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

	username := registration.Username
	accVal = &proto.AccountValue{
		Key:         &username,
		StringValue: &targetConf.Username,
	}
	accountValues = append(accountValues, accVal)

	password := registration.Password
	accVal = &proto.AccountValue{
		Key:         &password,
		StringValue: &targetConf.Password,
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
		var instanceMap map[string]map[string]struct{}
		for _, resource := range resources {
			if resource.Type == "aws_instance" {
				instanceMap = createNameToInstancesMap(resource)
			}
		}
		for name, v := range instanceMap {
			// Create Group when the instance size is > 1
			if len(v) > 1 {

			} else {
				for id := range v {
					entityDto, e := createEntityDto(name, id)
					if e != nil {
						glog.Errorf("Error building EntityDTO from metric %s", err)
						return nil, err
					}
					resultDTOs = append(resultDTOs, entityDto)
				}

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

func createNameToInstancesMap(resource *parser.Resource) map[string]map[string]struct{} {
	instanceNameToIntancesMap := make(map[string]map[string]struct{})
	name := resource.Name
	instanceSet, exist := instanceNameToIntancesMap[name]
	if !exist {
		instanceSet = make(map[string]struct{})
	}
	for _, instance := range resource.Instances {
		attributes := instance.Attributes
		if id, ok := attributes["id"]; ok {
			instanceSet[fmt.Sprintf("%v", id)] = struct{}{}
		}
	}
	instanceNameToIntancesMap[name] = instanceSet
	return instanceNameToIntancesMap
}

func createEntityDto(name string, id string) (*proto.EntityDTO, error) {
	entityDto, err := builder.NewEntityDTOBuilder(proto.EntityDTO_VIRTUAL_MACHINE, id).
		DisplayName(name).
		WithProperty(getEntityProperty("aws::us-east-2::VM::i-02661aa369e8b4ad9")).
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

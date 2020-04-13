package registration

import (
	"github.com/golang/glog"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	TargetIdField string = "targetIdentifier"
	propertyId    string = "id"
	Username      string = "username"
	Password      string = "password"
)

// Implements the TurboRegistrationClient interface
type TFRegistrationClient struct {
	TargetTypeSuffix string
}

func NewTFRegistrationClient(targetTypeSuffix string) (*TFRegistrationClient, error) {
	return &TFRegistrationClient{
		TargetTypeSuffix: targetTypeSuffix,
	}, nil
}

func (p *TFRegistrationClient) GetSupplyChainDefinition() []*proto.TemplateDTO {
	glog.Infoln("Building supply chain for Terraform Probe ..........")

	supplyChainFactory := &SupplyChainFactory{}
	supplyChain, err := supplyChainFactory.CreateSupplyChain()
	if err != nil {
		glog.Errorf("Failed to build supply chain: %v", err)
		return []*proto.TemplateDTO{}
	}
	return supplyChain
}

func (p *TFRegistrationClient) GetIdentifyingFields() string {
	return TargetIdField
}

func (rClient *TFRegistrationClient) GetAccountDefinition() []*proto.AccountDefEntry {
	var acctDefProps []*proto.AccountDefEntry

	// target ID
	targetIDAcctDefEntry := builder.NewAccountDefEntryBuilder(TargetIdField, "TargetID",
		"ID or address of the target", ".*", false, false).Create()
	acctDefProps = append(acctDefProps, targetIDAcctDefEntry)

	// username
	usernameAcctDefEntry := builder.NewAccountDefEntryBuilder(Username, "Username",
		"Username of the target cluster master", ".*", false, false).Create()
	acctDefProps = append(acctDefProps, usernameAcctDefEntry)

	// password
	passwordAcctDefEntry := builder.NewAccountDefEntryBuilder(Password, "Password",
		"Password of the target cluster master", ".*", false, true).Create()
	acctDefProps = append(acctDefProps, passwordAcctDefEntry)

	return acctDefProps
}

func (rClient *TFRegistrationClient) GetActionPolicy() []*proto.ActionPolicyDTO {
	glog.V(3).Infof("Begin to build Action Policies")
	ab := builder.NewActionPolicyBuilder()
	supported := proto.ActionPolicyDTO_SUPPORTED
	node := proto.EntityDTO_VIRTUAL_MACHINE
	nodePolicy := make(map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability)
	nodePolicy[proto.ActionItemDTO_RIGHT_SIZE] = supported

	rClient.addActionPolicy(ab, node, nodePolicy)

	return ab.Create()
}

func (rClient *TFRegistrationClient) addActionPolicy(ab *builder.ActionPolicyBuilder,
	entity proto.EntityDTO_EntityType,
	policies map[proto.ActionItemDTO_ActionType]proto.ActionPolicyDTO_ActionCapability) {

	for action, policy := range policies {
		ab.WithEntityActions(entity, action, policy)
	}
}

func (rclient *TFRegistrationClient) GetEntityMetadata() []*proto.EntityIdentityMetadata {

	glog.V(2).Infof("Begin to build EntityIdentityMetadata")

	result := []*proto.EntityIdentityMetadata{}

	entities := []proto.EntityDTO_EntityType{
		proto.EntityDTO_VIRTUAL_MACHINE,
	}

	for _, etype := range entities {
		meta := rclient.newIdMetaData(etype, []string{propertyId})
		result = append(result, meta)
	}

	glog.V(4).Infof("EntityIdentityMetaData: %++v", result)
	return result
}

func (rclient *TFRegistrationClient) newIdMetaData(etype proto.EntityDTO_EntityType, names []string) *proto.EntityIdentityMetadata {
	data := []*proto.EntityIdentityMetadata_PropertyMetadata{}
	for _, name := range names {
		dat := &proto.EntityIdentityMetadata_PropertyMetadata{
			Name: &name,
		}
		data = append(data, dat)
	}

	result := &proto.EntityIdentityMetadata{
		EntityType:            &etype,
		NonVolatileProperties: data,
	}

	return result
}

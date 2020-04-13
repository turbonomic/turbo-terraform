package registration

import (
	"github.com/enlinxu/turbo-terraform/pkg/discovery/constant"
	"github.com/turbonomic/turbo-go-sdk/pkg/builder"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"github.com/turbonomic/turbo-go-sdk/pkg/supplychain"
)

var (
	VMIPFieldPaths         = []string{constant.SUPPLY_CHAIN_CONSTANT_VIRTUAL_MACHINE_DATA}
	proxyVMUUID            = "Proxy_VM_UUID"
	VMUUID                 = supplychain.SUPPLY_CHAIN_CONSTANT_ID
	ActionEligibilityField = "actionEligibility"
)

type SupplyChainFactory struct{}

func (f *SupplyChainFactory) CreateSupplyChain() ([]*proto.TemplateDTO, error) {
	// VM node
	vmNode, err := f.buildVMSupplyBuilder()

	if err != nil {
		return nil, err
	}

	// Stitching metadata for the vm node
	vmMetadata, err := f.getVMStitchingMetaData()
	if err != nil {
		return nil, err
	}

	vmNode.MergedEntityMetaData = vmMetadata

	return supplychain.NewSupplyChainBuilder().
		Top(vmNode).
		Create()
}

func (f *SupplyChainFactory) buildVMSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_VIRTUAL_MACHINE)
	builder.SetPriority(-1)
	builder.SetTemplateType(proto.TemplateDTO_BASE)

	return builder.Create()
}

func (f *SupplyChainFactory) getVMStitchingMetaData() (*proto.MergedEntityMetadata, error) {

	var vmbuilder *builder.MergedEntityMetadataBuilder

	vmbuilder = builder.NewMergedEntityMetadataBuilder().
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		InternalMatchingProperty(proxyVMUUID).
		ExternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingField(VMUUID, []string{})

	metadata, err := vmbuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// Stitching metadata required for stitching with XL
func (f *SupplyChainFactory) buildNodeMergedEntityMetadata() (*proto.MergedEntityMetadata, error) {
	mergedEntityMetadataBuilder := builder.NewMergedEntityMetadataBuilder()

	mergedEntityMetadataBuilder.PatchField(ActionEligibilityField, []string{})
	// Set up matching criteria based on stitching type

	mergedEntityMetadataBuilder.
		InternalMatchingType(builder.MergedEntityMetadata_STRING).
		InternalMatchingProperty(proxyVMUUID).
		ExternalMatchingType(builder.MergedEntityMetadata_STRING).
		ExternalMatchingField(VMUUID, []string{})

	return mergedEntityMetadataBuilder.
		Build()
}
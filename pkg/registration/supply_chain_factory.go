package registration

import (
	"github.com/enlinxu/turbo-terraform/pkg/discovery/constant"
	"github.com/golang/glog"
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
	// Workload Controller supply chain template
	workloadControllerSupplyChainNode, err := f.buildWorkloadControllerSupplyBuilder()
	if err != nil {
		return nil, err
	}

	glog.V(4).Infof("Supply chain node: %+v", workloadControllerSupplyChainNode)
	return supplychain.NewSupplyChainBuilder().
		Top(vmNode).Entity(workloadControllerSupplyChainNode).
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
		InternalMatchingProperty(proxyVMUUID).
		ExternalMatchingField(VMUUID, []string{})

	metadata, err := vmbuilder.Build()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (f *SupplyChainFactory) buildWorkloadControllerSupplyBuilder() (*proto.TemplateDTO, error) {
	builder := supplychain.NewSupplyChainNodeBuilder(proto.EntityDTO_WORKLOAD_CONTROLLER)
	// Link from Pod to VM
	workloadControllerVMLink := supplychain.NewExternalEntityLinkBuilder()
	workloadControllerVMLink.Link(proto.EntityDTO_WORKLOAD_CONTROLLER, proto.EntityDTO_VIRTUAL_MACHINE, proto.Provider_LAYERED_OVER)

	workloadControllerLink, err := workloadControllerVMLink.Build()
	if err != nil {
		return nil, err
	}
	return builder.ConnectsTo(workloadControllerLink).Create()
}

package constant

const (
	// MetricType
	Used = "used"

	// Internal matching property
	// The default namespace of entity property
	DefaultPropertyNamespace = "DEFAULT"

	// The attribute used for stitching with other probes (e.g., prometurbo) with app and vapp
	StitchingAttr string = "IP"

	// External matching property
	// The attribute used for stitching with other probes (e.g., prometurbo) with vm
	SUPPLY_CHAIN_CONSTANT_IP_ADDRESS           string = "ipAddress"
	SUPPLY_CHAIN_CONSTANT_VIRTUAL_MACHINE_DATA        = "virtual_machine_data"
)

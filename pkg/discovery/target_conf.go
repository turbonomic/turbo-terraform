package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
)

const (
	// When user doesn't specify username and password, the default username and password will be used.
	defaultUsername      = "defaultUser"
	defaultPassword      = "defaultPassword"
	defaultProbeCategory = "Orchestrator"
	defaultTargetType    = "Terraform"
)

// Configuration Parameters to connect to an Target
type TargetConf struct {
	ProbeCategory string `json:"probeCategory,omitempty"`
	TargetType    string `json:"targetType,omitempty"`
	Identifier    string `json:"targetName,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
}

// Create a new ExampleClientConf from file. Other fields have default values and can be overridden.
func NewTargetConf(path string) (*TargetConf, error) {

	glog.Infof("[TargetConf] Read configuration from %s\n", path)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("failed to read file:%v", err.Error())
		return nil, err
	}

	var config TargetConf
	err = json.Unmarshal(file, &config)

	if err != nil {
		msg := fmt.Sprintf("Unmarshall error :%v\n", err)
		glog.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	config.ValidateK8sTargetConfig()
	glog.V(2).Infof("Results: %+v\n", config)

	return &config, nil
}

func (config *TargetConf) ValidateK8sTargetConfig() error {
	if config.Identifier == "" {
		return fmt.Errorf("targetIdentifier is not provided")
	}

	// Prefix target id (address) with the target type (i.e., "Kubernetes-") to
	// avoid duplicate target id with other types of targets (e.g., aws).
	config.Identifier = defaultTargetType + "-" + config.Identifier

	if config.Username == "" {
		config.Username = defaultUsername
	}

	if config.Password == "" {
		config.Password = defaultPassword
	}

	if config.ProbeCategory == "" {
		config.ProbeCategory = defaultProbeCategory
	}

	if config.TargetType == "" {
		config.TargetType = defaultTargetType
	}
	return nil
}

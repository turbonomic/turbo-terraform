package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type TFState struct {
	Version   int         `json:"version,omitempty"`
	Resources []*Resource `json:"resources,omitempty"`
}

type Resource struct {
	Type      string      `json:"type,omitempty"`
	Name      string      `json:"name,omitempty"`
	Provider  string      `json:"provider,omitempty"`
	Instances []*Instance `json:"instances,omitempty"`
}

type Instance struct {
	IndexKey   int                    `json:"index_key,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Parse the Terraform .tfstate file at the given path
func ParseTerraformStateFile(path string) (*TFState, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("File error: %v\n" + err.Error())
	}

	terraformState := &TFState{}

	if err := json.Unmarshal(bytes, terraformState); err != nil {
		return nil, fmt.Errorf("Unmarshall error :%v", err.Error())
	}

	return terraformState, nil
}

// Parse the Terraform .tfstate file at the given path
func ParseTerraformStateAPI(bytes []byte) (*TFState, error) {
	terraformState := &TFState{}

	if err := json.Unmarshal(bytes, terraformState); err != nil {
		return nil, fmt.Errorf("Unmarshall error :%v", err.Error())
	}

	return terraformState, nil
}

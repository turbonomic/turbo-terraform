package query

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type QueryResultInfo struct {
	Data *DataInfo `json:"data,omitempty"`
}

type QueryResultInfoArray struct {
	Data []*DataInfo `json:"data,omitempty"`
}

type DataInfo struct {
	Id         string                 `json:"id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func GetWorkspaces(org string, token string) ([]byte, error) {
	url := "https://app.terraform.io/api/v2/organizations/" + org + "/workspaces"
	method := "GET"
	body, err := Query(url, method, nil, token)
	if err != nil {
		glog.Errorf("Error getting Workspaces", err)
		return nil, err
	}
	return body, nil
}

func GetCurrentStateVersion(workspaceId string, token string) ([]byte, error) {
	url := "https://app.terraform.io/api/v2/workspaces/" + workspaceId + "/current-state-version"
	method := "GET"
	body, err := Query(url, method, nil, token)
	if err != nil {
		glog.Errorf("Error getting CurrentStateVersionResult", err)
		return nil, err
	}
	var stateResult QueryResultInfo
	err = json.Unmarshal(body, &stateResult)
	if err != nil {
		glog.Errorf("Error unmarshalling CurrentStateVersionResult", err)
		return nil, err
	}
	hostURL := fmt.Sprintf("%v", stateResult.Data.Attributes["hosted-state-download-url"])
	hostResponse, err := Query(hostURL, method, nil, token)
	if err != nil {
		glog.Errorf("Error getting CurrentStateVersionResult", err)
		return nil, err
	}
	return hostResponse, nil
}

func GetVariables(org string, workspace string, token string) ([]byte, error) {
	url := "https://app.terraform.io/api/v2/vars?filter%5Borganization%5D%5Bname%5D=" + org + "&filter%5Bworkspace%5D%5Bname%5D=" + workspace
	method := "GET"
	body, err := Query(url, method, nil, token)
	if err != nil {
		glog.Errorf("Error getting Workspaces", err)
		return nil, err
	}
	return body, nil
}

func UpdateVariables(varId string, size string, token string) ([]byte, error) {
	url := "https://app.terraform.io/api/v2/vars/" + varId
	method := "PATCH"
	payload := strings.NewReader("{\n  \"data\": {\n    \"id\":\"" + varId + "\",\n    \"attributes\": {\n      \"key\":\"instance_type\",\n      \"value\":\"" + size + "\",\n      \"category\":\"terraform\",\n      \"hcl\": false,\n      \"sensitive\": false\n    },\n    \"type\":\"vars\"\n  }\n}")
	body, err := Query(url, method, payload, token)
	if err != nil {
		glog.Errorf("Error getting Workspaces", err)
		return nil, err
	}
	return body, nil
}

func Query(url string, method string, payload io.Reader, token string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		glog.Errorf("Error building HTTP request", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/vnd.api+json")
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error submitting HTTP request", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Error reading response", err)
		return nil, err
	}
	return body, nil
}

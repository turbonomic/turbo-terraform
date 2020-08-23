package executor

import (
	"github.com/enlinxu/turbo-terraform/pkg/discovery"
	query "github.com/enlinxu/turbo-terraform/pkg/query"
	"github.com/golang/glog"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"io/ioutil"
	"strings"
)

type scaleActionExecutor struct {
	Name    string
	tfToken *string
}

func NewScaleActionExecutor(name string, tfToken *string) *scaleActionExecutor {
	return &scaleActionExecutor{
		Name:    name,
		tfToken: tfToken,
	}
}

func (e *scaleActionExecutor) Execute(actionItem *proto.ActionItemDTO, progressTracker sdkprobe.ActionProgressTracker) error {
	glog.V(2).Infof("[%v] begin to execute action ...", e.Name)
	vmName := actionItem.GetTargetSE().Id
	targetProfileName := actionItem.GetNewSE().DisplayName
	currentProfileName := actionItem.GetCurrentSE().DisplayName
	files := discovery.EntityIdToAssetsMap[*vmName]
	for file := range files {
		if strings.HasPrefix(file, "var") {
			// TF Enterprise : The assets here is the variable
			result, err := query.UpdateVariables(file, *targetProfileName, *e.tfToken)
			if err != nil {
				glog.Error("Failed to update variables %v" + err.Error())
			}
			glog.V(2).Info(string(result))
		} else if strings.Index(file, ".tfstate") > -1 {
			//Ignore the tfstate file
			continue
		} else {
			// TF Community Edition. This implementation is just a PoC implementation, needs more discussion
			read, err := ioutil.ReadFile(file)
			if err != nil {
				glog.Error("Failed to read files %v" + err.Error())
				continue
			}
			//fmt.Println(string(read))
			if strings.Index(string(read), *currentProfileName) > -1 {
				newContents := strings.Replace(string(read), *currentProfileName, *targetProfileName, -1)

				glog.V(2).Info(newContents)

				err = ioutil.WriteFile(file, []byte(newContents), 0)
				if err != nil {
					glog.Error("Failed to read files %v" + err.Error())
					return err
				}
			}
		}
	}
	glog.V(2).Infof("[%v] executing action:\n%++v", e.Name, actionItem)
	//TODO: update progress

	glog.V(2).Infof("[%v] end of executing action ...", e.Name)
	return nil
}

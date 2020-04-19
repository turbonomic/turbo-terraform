package action

import (
	"github.com/enlinxu/turbo-terraform/pkg/discovery"
	"github.com/golang/glog"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"io/ioutil"
	"strings"
)

type generalActionExecutor struct {
	Name string
}

func NewActionExecutor(name string) *generalActionExecutor {
	return &generalActionExecutor{
		Name: name,
	}
}

func (e *generalActionExecutor) Execute(actionItem *proto.ActionItemDTO, progressTracker sdkprobe.ActionProgressTracker) error {
	glog.V(2).Infof("[%v] begin to execute action ...", e.Name)
	vmName := actionItem.GetTargetSE().Id
	targetProfileName := actionItem.GetNewSE().DisplayName
	currentProfileName := actionItem.GetCurrentSE().DisplayName
	files := discovery.EntityIdToFilesMap[*vmName]
	for file := range files {
		//Ignore the tfstate file
		if strings.Index(file, ".tfstate") > -1 {
			continue
		}
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
	glog.V(2).Infof("[%v] executing action:\n%++v", e.Name, actionItem)
	//TODO: update progress

	glog.V(2).Infof("[%v] end of executing action ...", e.Name)
	return nil
}

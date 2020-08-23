package executor

import (
	"fmt"
	"github.com/enlinxu/turbo-terraform/pkg/discovery"
	"github.com/golang/glog"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"io/ioutil"
	"strings"
)

type resizeActionExecutor struct {
	Name string
}

func NewResizeActionExecutor(name string) *resizeActionExecutor {
	return &resizeActionExecutor{
		Name: name,
	}
}

func (e *resizeActionExecutor) Execute(actionItem *proto.ActionItemDTO, progressTracker sdkprobe.ActionProgressTracker) error {
	glog.V(2).Infof("[%v] begin to execute action ...", e.Name)
	vmName := actionItem.GetTargetSE().Id
	glog.V(2).Infof("The vm name is [%v]", vmName)
	comm1 := actionItem.GetCurrentComm()
	comm2 := actionItem.GetNewComm()

	if comm1.GetCommodityType() != comm2.GetCommodityType() {
		return fmt.Errorf("commodity type does not match %v vs %v",
			comm1.CommodityType.String(), comm2.CommodityType.String())
	}
	files := discovery.EntityIdToAssetsMap[*vmName]
	for file := range files {
		glog.V(2).Infof("The file name is [%v]", file)
		//Ignore the tfstate file
		if strings.Index(file, ".tfstate") > -1 {
			continue
		}
		read, err := ioutil.ReadFile(file)
		if err != nil {
			glog.Error("Failed to read files %v" + err.Error())
			continue
		}
		fmt.Println(string(read))
	}
	glog.V(2).Infof("[%v] executing action:\n%++v", e.Name, actionItem)
	//TODO: update progress

	glog.V(2).Infof("[%v] end of executing action ...", e.Name)
	return nil
}

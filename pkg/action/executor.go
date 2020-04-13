package action

import (
	"github.com/golang/glog"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
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

	glog.V(2).Infof("[%v] executing action:\n%++v", e.Name, actionItem)
	//TODO: update progress

	glog.V(2).Infof("[%v] end of executing action ...", e.Name)
	return nil
}

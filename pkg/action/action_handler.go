package action

import (
	"fmt"
	"github.com/golang/glog"
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
	"time"
)

type ActionHandler struct {
	actionExecutors map[TurboActionType]TurboExecutor
	stop            chan struct{}
}

func NewActionHandler(stop chan struct{}) *ActionHandler {
	executors := make(map[TurboActionType]TurboExecutor)

	handler := &ActionHandler{
		stop:            stop,
		actionExecutors: executors,
	}

	handler.registerExecutors()
	return handler
}

func (h *ActionHandler) String() string {

	atypes := []TurboActionType{}
	for k := range h.actionExecutors {
		atypes = append(atypes, k)
	}

	return fmt.Sprintf("%v", atypes)
}

func (h *ActionHandler) registerExecutors() {
	vmResizer := NewActionExecutor("vmResize")
	h.actionExecutors[ActionResizeVM] = vmResizer
}

func (h *ActionHandler) goodResult(msg string) *proto.ActionResult {

	state := proto.ActionResponseState_SUCCEEDED
	progress := int32(100)

	res := &proto.ActionResponse{
		ActionResponseState: &state,
		Progress:            &progress,
		ResponseDescription: &msg,
	}

	return &proto.ActionResult{
		Response: res,
	}
}

func (h *ActionHandler) failedResult(msg string) *proto.ActionResult {

	state := proto.ActionResponseState_FAILED
	progress := int32(0)

	res := &proto.ActionResponse{
		ActionResponseState: &state,
		Progress:            &progress,
		ResponseDescription: &msg,
	}

	return &proto.ActionResult{
		Response: res,
	}
}

func (h *ActionHandler) ExecuteAction(
	actionDTO *proto.ActionExecutionDTO,
	accountValue []*proto.AccountValue,
	progressTracker sdkprobe.ActionProgressTracker) (*proto.ActionResult, error) {

	actionItems := actionDTO.GetActionItem()
	action := actionItems[0]

	glog.V(3).Infof("action:%+++v", action)
	actionType, err := getActionType(action)
	if err != nil {
		msg := fmt.Sprintf("failed to get Action Type:%v", err.Error())
		glog.Error(msg)
		result := h.failedResult(msg)
		return result, nil
	}

	executor, exist := h.actionExecutors[actionType]
	if !exist {
		msg := fmt.Sprintf("action type [%v] is not supported", actionType)
		glog.Error(msg)
		result := h.failedResult(msg)
		return result, nil
	}

	// here progressTracker is used to keep alive; executor won't really use it.
	stop := make(chan struct{})
	defer close(stop)
	go keepAlive(progressTracker, stop)

	err = executor.Execute(action, progressTracker)
	if err != nil {
		msg := fmt.Sprintf("Action failed: %v", err.Error())
		glog.Error(msg)
		result := h.failedResult(msg)
		return result, nil
	}

	result := h.goodResult("Success")
	return result, nil
}

func getActionType(action *proto.ActionItemDTO) (TurboActionType, error) {
	atype := action.GetActionType()
	object := action.GetTargetSE()
	if object == nil {
		return ActionUnknown, fmt.Errorf("TargetSE is empty.")
	}
	objectType := object.GetEntityType()

	glog.V(2).Infof("action [%v-%v] is received.", atype, objectType)

	switch atype {
	case proto.ActionItemDTO_RIGHT_SIZE:
		glog.V(4).Infof("[%v] [%v]", atype, objectType)
		switch objectType {
		case proto.EntityDTO_VIRTUAL_MACHINE:
			return ActionResizeVM, nil
		}
	}

	err := fmt.Errorf("Action [%v-%v] is not supported.", atype, objectType)
	return ActionUnknown, err
}

func keepAlive(tracker sdkprobe.ActionProgressTracker, stop chan struct{}) {
	//TODO: add timeout
	go func() {
		var progress int32 = 0
		state := proto.ActionResponseState_IN_PROGRESS

		for {
			progress = progress + 1
			if progress > 99 {
				progress = 99
			}

			tracker.UpdateProgress(state, "in progress", progress)

			t := time.NewTimer(time.Second * 3)
			select {
			case <-stop:
				glog.V(3).Infof("action keepAlive goroutine exit.")
				return
			case <-t.C:
			}
		}
	}()
}

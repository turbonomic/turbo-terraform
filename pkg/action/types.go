package action

import (
	sdkprobe "github.com/turbonomic/turbo-go-sdk/pkg/probe"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type TurboActionType string

const (
	ActionResizeVM TurboActionType = "resizeVirtualMachine"
	ActionUnknown  TurboActionType = "unknown"
)

type TurboExecutor interface {
	Execute(actionItem *proto.ActionItemDTO, progressTracker sdkprobe.ActionProgressTracker) error
}

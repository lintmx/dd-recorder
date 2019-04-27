package instance

import (
	"context"
	"github.com/lintmx/dd-recorder/configs"
	"sync"
)

// InstanceKey for Context
const InstanceKey string = "meaqua"

// Instance struct
type Instance struct {
	WaitGroup *sync.WaitGroup
	Config    *configs.Config
}

// GetInstance get ctx instance
func GetInstance(ctx context.Context) *Instance {
	// Get Instance and type assertion
	if inst, ok := ctx.Value(InstanceKey).(*Instance); ok {
		return inst
	}
	return nil
}

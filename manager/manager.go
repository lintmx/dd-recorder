package manager

import (
	"context"
	"github.com/lintmx/dd-recorder/api"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/monitor"
	"github.com/lintmx/dd-recorder/utils"
	"go.uber.org/zap"
	"net/url"
	"time"
)

// DD start
func DD(ctx context.Context) {
	inst := instance.GetInstance(ctx)

	// run monitor with room
	for _, room := range inst.Config.Rooms {
		u, err := url.Parse(room)

		if err != nil {
			inst.Logger.Error("Room Url Parse Error", zap.String("url", room))
			continue
		}

		api := api.Check(u)
		if api == nil {
			inst.Logger.Error("Room not support", zap.String("host", u.Host))
		} else {
			inst.WaitGroup.Add(1)
			m := monitor.Monitor{
				MonitorID:  utils.GetMd5(u.String()),
				LiveAPI:    api,
				TimeTicker: time.NewTicker(time.Duration(inst.Config.Interval) * time.Second),
			}
			go m.Run(ctx)
		}
	}
}

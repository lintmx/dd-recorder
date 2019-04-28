package manager

import (
	"context"
	"github.com/lintmx/dd-recorder/api"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/monitor"
	"github.com/lintmx/dd-recorder/utils"
	"go.uber.org/zap"
	"net/url"
)

// DD start
func DD(ctx context.Context) {
	inst := instance.GetInstance(ctx)

	// run monitor with room
	for _, room := range inst.Config.Rooms {
		u, err := url.Parse(room)

		if err != nil {
			zap.S().Error("Room Url Parse Error", zap.String("url", room))
			continue
		}

		api := api.Check(u)
		if api == nil {
			zap.S().Error("Room not support", zap.String("host", u.Host))
		} else {
			inst.WaitGroup.Add(1)
			m := monitor.Monitor{
				MonitorID: utils.BKDRHash64(u.String()),
				LiveAPI:   api,
			}

			zap.L().Info("Monitor Init",
				zap.String("Id", m.MonitorID),
				zap.String("Author", m.LiveAPI.GetAuthor()),
				zap.String("Platform", m.LiveAPI.GetPlatformName()),
			)
			go m.Run(ctx)
		}
	}
}

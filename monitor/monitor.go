package monitor

import (
	"context"
	"github.com/lintmx/dd-recorder/api"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/record"
	"go.uber.org/zap"
	"time"
)

// Monitor struct
type Monitor struct {
	MonitorID  string
	LiveAPI    api.LiveAPI
	LiveStatus bool
	StopChan   chan struct{}
	rec        *record.Record
}

// Run a dd monitor
func (m *Monitor) Run(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	defer inst.WaitGroup.Done()
	timer := time.NewTimer(0)
	m.rec = record.New(m.MonitorID, m.LiveAPI)

	for {
		select {
		case <-ctx.Done(): // Exit Signal
			m.rec.Stop()
			return
		case <-timer.C:
			m.refresh(ctx)
			timer.Reset(time.Duration(inst.Config.Interval) * time.Second)
		}
	}
}

// Refresh live status
func (m *Monitor) refresh(ctx context.Context) {
	err := m.LiveAPI.RefreshLiveInfo()

	if err != nil {
		zap.L().Error("Refresh Live Info",
			zap.String("MonitorId", m.MonitorID),
			zap.String("Err", err.Error()),
		)
		return
	}

	if m.LiveStatus != m.LiveAPI.GetLiveStatus() {
		m.LiveStatus = m.LiveAPI.GetLiveStatus()

		if m.LiveStatus {
			instance.GetInstance(ctx).WaitGroup.Add(1)
			go m.rec.Start(ctx)
		} else {
			m.rec.Stop()
		}
	}
}

package monitor

import (
	"context"
	"fmt"
	"github.com/lintmx/dd-recorder/api"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/record"
	"github.com/lintmx/dd-recorder/utils"
	"go.uber.org/zap"
	"time"
)

// Monitor struct
type Monitor struct {
	MonitorID  string
	LiveAPI    api.LiveAPI
	TimeTicker *time.Ticker
	LiveStatus bool
	StopChan   chan struct{}
	rec        record.Record
}

// Run a dd monitor
func (m *Monitor) Run(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	inst.Logger.Info("Init Monitor",
		zap.String("MonitorId", m.MonitorID),
		zap.String("Platform", m.LiveAPI.GetPlatformName()),
		zap.String("Url", m.LiveAPI.GetLiveURL()),
	)
	defer inst.WaitGroup.Done()
	defer inst.Logger.Info("Stop Monitor",
		zap.String("MonitorId", m.MonitorID),
		zap.String("Platform", m.LiveAPI.GetPlatformName()),
		zap.String("Url", m.LiveAPI.GetLiveURL()),
	)

	m.rec = record.Record{
		MonitorID: m.MonitorID,
		LiveAPI:   m.LiveAPI,
	}

	m.refresh(ctx)
	for {
		select {
		case <-ctx.Done():
			// turn off started record
			if m.LiveStatus == true {
				m.stop()
			}
			return
		case <-m.TimeTicker.C:
			m.refresh(ctx)
		}
	}
}

// Refresh live status
func (m *Monitor) refresh(ctx context.Context) {
	err := m.LiveAPI.RefreshLiveInfo()

	if err != nil {
		instance.GetInstance(ctx).Logger.Error(err.Error(),
			zap.String("MonitorId", m.MonitorID),
			zap.String("Platform", m.LiveAPI.GetPlatformName()),
			zap.String("Url", m.LiveAPI.GetLiveURL()),
		)
		return
	}

	if m.LiveStatus != m.LiveAPI.GetLiveStatus() {
		m.LiveStatus = m.LiveAPI.GetLiveStatus()

		if m.LiveStatus {
			m.start(ctx)
		} else {
			m.stop()
		}
	}
}

// start a record
func (m *Monitor) start(ctx context.Context) {
	instance.GetInstance(ctx).WaitGroup.Add(1)
	m.rec.StopChan = make(chan struct{})
	m.rec.RecordID = utils.GetMd5(fmt.Sprintf("%s%s", m.MonitorID, m.LiveAPI.GetTitle()))
	go m.rec.Run(ctx)
}

// stop a record
func (m *Monitor) stop() {
	m.rec.Stop()
}

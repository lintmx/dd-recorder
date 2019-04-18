package record

import (
	"context"
	"fmt"
	"github.com/lintmx/dd-recorder/api"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/utils"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Record struct
type Record struct {
	MonitorID string
	RecordID  string
	LiveAPI   api.LiveAPI
	StopChan  chan struct{}
	cmd       *exec.Cmd
}

// Run record
func (r *Record) Run(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	inst.Logger.Info("Start Record",
		zap.String("MonitorId", r.MonitorID),
		zap.String("RecordId", r.RecordID),
		zap.String("Author", r.LiveAPI.GetAuthor()),
	)
	defer inst.WaitGroup.Done()
	defer inst.Logger.Info("Stop Record",
		zap.String("Id", r.MonitorID),
		zap.String("RecordId", r.RecordID),
		zap.String("Author", r.LiveAPI.GetAuthor()),
	)

	for {
		select {
		case <-r.StopChan:
			return
		default:
			streamURLs, err := r.LiveAPI.GetStreamURLs()
			now := time.Now()

			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}

			streamURL := api.StreamURL{}
			for _, stream := range streamURLs {
				// TODO: 播放链接选择
				streamURL = stream
				break
			}

			if streamURL == (api.StreamURL{}) {
				time.Sleep(3 * time.Second)
				continue
			}

			outPath := filepath.Join(inst.Config.OutPath,
				utils.FilterInvalidCharacters(r.LiveAPI.GetPlatformName()),
				utils.FilterInvalidCharacters(r.LiveAPI.GetAuthor()),
				now.Format("2006-01-02"),
			)
			os.MkdirAll(outPath, os.ModePerm)
			outFile := filepath.Join(outPath,
				fmt.Sprintf("[%s][%s][%s] %s.%s",
					now.Format("2006-01-02 15:04:05"),
					utils.FilterInvalidCharacters(r.LiveAPI.GetPlatformName()),
					utils.FilterInvalidCharacters(r.LiveAPI.GetAuthor()),
					utils.FilterInvalidCharacters(r.LiveAPI.GetTitle()),
					streamURL.FileType,
				),
			)

			r.cmd = exec.Command("ffmpeg",
				"-y",
				"-i", streamURL.PlayURL.String(),
				"-c", "copy",
				outFile,
			)

			r.cmd.Start()
			r.cmd.Wait()
		}
	}
}

// Stop record
func (r *Record) Stop() {
	close(r.StopChan)
	r.cmd.Process.Kill()
}

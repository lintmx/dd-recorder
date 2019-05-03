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
	"sync"
	"time"
)

// Record struct
type Record struct {
	MonitorID    string
	RecordID     string
	RecordStatus bool
	doneChan     chan struct{}
	LiveAPI      api.LiveAPI
	outPath      string
	outFile      string
	startTime    time.Time
	cmd          *exec.Cmd
	waitGroup    *sync.WaitGroup
}

// New and return a Record
func New(monitorID string, liveAPI api.LiveAPI) *Record {
	record := Record{
		MonitorID: monitorID,
		LiveAPI:   liveAPI,
		waitGroup: &sync.WaitGroup{},
	}

	return &record
}

// Start Record
func (r *Record) Start(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	defer inst.WaitGroup.Done()
	r.doneChan = make(chan struct{})
	if r.RecordStatus {
		return
	}
	r.RecordStatus = true
	r.doneChan = make(chan struct{})
	r.outPath = filepath.Join(inst.Config.OutPath,
		utils.FilterInvalidCharacters(r.LiveAPI.GetPlatformName()),
		utils.FilterInvalidCharacters(r.LiveAPI.GetAuthor()),
		time.Now().Format("2006-01-02"),
	)
	os.MkdirAll(r.outPath, os.ModePerm)

	zap.L().Info("Record Start",
		zap.String("Id", r.MonitorID),
		zap.String("Author", r.LiveAPI.GetAuthor()),
		zap.String("Title", r.LiveAPI.GetTitle()),
	)

	r.waitGroup.Add(1)
	go r.recordStream()
	r.waitGroup.Add(1)
	go r.recordDanmaku()
	r.waitGroup.Wait()
	r.RecordStatus = false
	r.outPath = ""
	r.outFile = ""

	zap.L().Info("Record Stop",
		zap.String("Id", r.MonitorID),
		zap.String("Author", r.LiveAPI.GetAuthor()),
		zap.String("Title", r.LiveAPI.GetTitle()),
	)
}

func (r *Record) recordStream() {
	defer r.waitGroup.Done()
	for {
		select {
		case <-r.doneChan:
			return
		default:
			streamURLs, err := r.LiveAPI.GetStreamURLs()
			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}

			streamURL := api.StreamURL{}
			for _, stream := range streamURLs {
				// TODO: Stream Url Select
				streamURL = stream
				break
			}

			if streamURL == (api.StreamURL{}) {
				time.Sleep(3 * time.Second)
				continue
			}
			t := time.Now()

			r.startTime = t
			r.outFile = filepath.Join(r.outPath,
				fmt.Sprintf("[%s][%s][%s] %s",
					t.Format("2006-01-02 15-04-05"),
					utils.FilterInvalidCharacters(r.LiveAPI.GetPlatformName()),
					utils.FilterInvalidCharacters(r.LiveAPI.GetAuthor()),
					utils.FilterInvalidCharacters(r.LiveAPI.GetTitle()),
				),
			)

			r.cmd = exec.Command("ffmpeg",
				"-loglevel", "warning",
				"-y",
				"-timeout", "30000000",
				"-i", streamURL.PlayURL.String(),
				"-c", "copy",
				fmt.Sprintf("%s.%s", r.outFile, streamURL.FileType),
			)

			r.cmd.Start()
			r.cmd.Wait()
		}
	}
}

func (r *Record) recordDanmaku() {
	defer r.waitGroup.Done()
	msg, err := r.LiveAPI.GetDanmaku(r.doneChan)
	if err != nil {
		return
	}
	for r.outFile == "" {
		time.Sleep(500 * time.Millisecond)
	}

	file, err := os.OpenFile(
		fmt.Sprintf("%s.%s", r.outFile, "xml"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0755,
	)
	if err != nil {
		return
	}

	lastFileName := r.outFile
	file.WriteString(
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?><i><chatserver>chat.bilibili.com</chatserver><chatid>0</chatid><mission>0</mission><maxlimit>0</maxlimit><source>k-v</source>\n",
	)
	for m := range msg {
		if lastFileName != r.outFile {
			file.WriteString("</i>")
			file.Close()
			file, err = os.OpenFile(
				fmt.Sprintf("%s.%s", r.outFile, "xml"),
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0666,
			)
			if err != nil {
				continue
			}
			lastFileName = r.outFile
			file.WriteString(
				"<?xml version=\"1.0\" encoding=\"UTF-8\"?><i><chatserver>chat.bilibili.com</chatserver><chatid>0</chatid><mission>0</mission><maxlimit>0</maxlimit><source>k-v</source>\n",
			)
		}

		// TODO: fix negative number
		file.WriteString(
			fmt.Sprintf(
				"<d p=\"%.3f,1,25,16777215,%d,0,0,0\">%s</d>\n",
				time.Now().Sub(r.startTime).Seconds(),
				m.SendTime,
				m.Content,
			),
		)
	}
	file.WriteString("</i>")
	file.Close()
}

// Stop record
func (r *Record) Stop() {
	if r.RecordStatus {
		close(r.doneChan)
		r.cmd.Process.Kill()
		r.RecordStatus = false
	}
}

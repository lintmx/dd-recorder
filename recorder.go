package main

import (
	"sync"
	"context"
	"fmt"
	"github.com/lintmx/dd-recorder/configs"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/logger"
	"github.com/lintmx/dd-recorder/manager"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// App Variable
var (
	Name      string
	Version   string
	Build     string
	GoVersion string
)

// Flag Variable
var (
	help     bool
	conf     string
	version  bool
	path     string
	rooms    []string
	interval uint16
	logPath  string
	debug    bool
)

func init() {
	flag.BoolVarP(&help, "help", "h", false, "This help")
	flag.StringVarP(&conf, "config", "c", "", "Configuration file")
	flag.BoolVarP(&version, "version", "v", false, "Show version number")
	flag.StringVarP(&path, "out_path", "o", "./Lives", "Live Video output path")
	flag.StringArrayVarP(&rooms, "url", "u", nil, "Live Rooms url")
	flag.Uint16Var(&interval, "interval", 10, "Refresh second")
	flag.StringVar(&logPath, "log", "", "Log Path")
	flag.BoolVar(&debug, "debug", false, "Debug Mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage of %s:\n", Name)
		flag.PrintDefaults()
	}

	flag.Parse()
}

func main() {
	if help {
		flag.Usage()
		os.Exit(0)
	}

	if version {
		fmt.Fprintf(os.Stdout, "%s (%s) Build in %s With %s\n", Name, Version, Build, GoVersion)
		os.Exit(0)
	}

	// Check FFmpeg
	if _, ok := exec.LookPath("ffmpeg"); ok != nil {
		fmt.Fprintf(os.Stdout, "[Error] FFmpeg not found.\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "DD recorder - 誰でも大好き！\n\n")

	var config *configs.Config

	// Init Config
	if conf != "" {
		config = configs.InitConfig(conf)
	} else {
		config = &configs.Config{
			Interval: interval,
			OutPath:  path,
			Rooms:    rooms,
			LogPath:  logPath,
			Debug:    debug,
		}
	}

	// Init Logger
	log := logger.InitLogger(config.Debug, config.LogPath)
	defer log.Sync()
	zap.ReplaceGlobals(log)

	inst := &instance.Instance{
		Config:    config,
		WaitGroup: &sync.WaitGroup{},
	}
	ctx := context.WithValue(context.Background(), instance.InstanceKey, inst)
	ctx, cannel := context.WithCancel(ctx)

	// start dd
	manager.DD(ctx)

	// Catch the exit signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-sigCh
		cannel()
	}()

	inst.WaitGroup.Wait()
	fmt.Fprintf(os.Stdout, "\nさようなら～\n")
}

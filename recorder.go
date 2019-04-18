package main

import (
	"context"
	"fmt"
	"github.com/lintmx/dd-recorder/configs"
	"github.com/lintmx/dd-recorder/instance"
	"github.com/lintmx/dd-recorder/manager"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// App Variable
var (
	Name      string
	Version   string
	Build     string
	GoVersion string
)

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
	flag.StringVar(&logPath, "log", "", "Log File")
	flag.BoolVar(&debug, "debug", false, "Debug Mode")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage of %s:\n", Name)
		flag.PrintDefaults()
	}

	flag.Parse()
}

func initLogger(levelAtom zap.AtomicLevel, logPath string) *zap.Logger {
	stdOut := []string{"stdout"}
	stdErr := []string{"stderr"}

	if logPath != "" {
		stdOut = append(stdOut, logPath)
		stdErr = append(stdErr, logPath)
	}

	config := zap.Config{
		Level:       levelAtom,
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			TimeKey:     "time",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
		OutputPaths:      stdOut,
		ErrorOutputPaths: stdErr,
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("log init error: %v", err))
	}

	return logger
}

func main() {
	if help {
		flag.Usage()
		os.Exit(0)
	}

	if version {
		fmt.Fprintf(os.Stdout, "%s (%s) Build %s With %s\n", Name, Version, Build, GoVersion)
		os.Exit(0)
	}

	// Check FFmpeg
	if _, ok := exec.LookPath("ffmpeg"); ok != nil {
		fmt.Fprintf(os.Stdout, "[Error] FFmpeg not found.\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%s - %s 誰でも大好き！\n", Name, Version)

	config := &configs.Config{}

	// Init Config
	if conf != "" {
		if !filepath.IsAbs(conf) {
			currentDir, _ := os.Getwd()
			conf = filepath.Join(currentDir, conf)
		}

		if c, err := configs.Parse(conf); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		} else {
			config = c
		}
	} else {
		config = &configs.Config{
			Interval: interval,
			OutPath:  path,
			Rooms:    rooms,
			LogPath:  logPath,
			Debug:    debug,
		}
	}

	levelAtom := zap.NewAtomicLevel()

	if config.Debug {
		levelAtom.SetLevel(zap.DebugLevel)
	}

	inst := &instance.Instance{
		Config: config,
		Logger: initLogger(levelAtom, config.LogPath),
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

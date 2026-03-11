package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/XANi/go-yamlcfg"
	"github.com/XANi/prometheus-modbus-exporter/config"
	"github.com/XANi/prometheus-modbus-exporter/modbus_client"
	"github.com/XANi/prometheus-modbus-exporter/web"
	"github.com/efigence/go-mon"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/fs"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var version string
var log *zap.SugaredLogger
var debug = false

// /* embeds with all files, just dir/ ignores files starting with _ or .
//
//go:embed static templates
var embeddedWebContent embed.FS

func init() {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	// naive systemd detection. Drop timestamp if running under it
	if os.Getenv("JOURNAL_STREAM") != "" {
		consoleEncoderConfig.TimeKey = ""
	}
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return (lvl < zapcore.ErrorLevel) != (lvl == zapcore.DebugLevel && !debug)
	})
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, os.Stderr, lowPriority),
		zapcore.NewCore(consoleEncoder, os.Stderr, highPriority),
	)
	logger := zap.New(core)
	if debug {
		logger = logger.WithOptions(
			zap.Development(),
			zap.AddCaller(),
			zap.AddStacktrace(highPriority),
		)
	} else {
		logger = logger.WithOptions(
			zap.AddCaller(),
		)
	}
	log = logger.Sugar()

}

func main() {
	defer log.Sync()
	// register internal stats
	mon.RegisterGcStats()
	app := &cli.Command{
		Name:        "prometheus-modbus-exporter",
		Aliases:     nil,
		Usage:       "",
		UsageText:   "",
		ArgsUsage:   "",
		Version:     version,
		Description: "prometheus modbus exporter",
		Flags:       nil,
		Commands:    nil,
		HideHelp:    true,
	}
	log.Infof("Starting %s version: %s", app.Name, version)
	app.Flags = []cli.Flag{
		&cli.BoolFlag{Name: "help, h", Usage: "show help"},
		&cli.BoolFlag{Name: "debug, d", Usage: "enable debug logs"},
		&cli.StringFlag{
			Name:  "listen-addr",
			Value: "127.0.0.1:3001",
			Usage: "Listen addr",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LISTEN_ADDR"),
			),
		},
		&cli.StringFlag{
			Name:  "pprof-addr",
			Value: "",
			Usage: "address to run pprof on, disabled by default",
		},
	}
	app.Action = func(ctx context.Context, c *cli.Command) error {
		if c.Bool("help") {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		debug = c.Bool("debug")
		log.Debug("debug enabled")

		cfgFiles := []string{
			"$HOME/.config/prometheus-modbus-exporter/config.yaml",
			"/etc/prometheus-modbus-exporter/config.yaml",
		}
		var cfg config.Config
		err := yamlcfg.LoadConfig(cfgFiles, &cfg)
		for bus, buscfg := range cfg.Bus {
			buscfg.Name = bus
			log.Infof("initializing bus %s", bus)
			_, err := modbus_client.New(modbus_client.Config{
				Bus:           buscfg,
				PrometheusURL: cfg.PrometheusURL,
				Logger:        log.Named(fmt.Sprintf("bus-%s", bus)),
			})
			if err != nil {
				log.Errorf("error initializing bus %s: %s", bus, err)
			}
		}

		var webDir fs.FS
		webDir = embeddedWebContent
		if st, err := os.Stat("./static"); err == nil && st.IsDir() {
			if st, err := os.Stat("./templates"); err == nil && st.IsDir() {
				webDir = os.DirFS(".")
				log.Infof(`detected directories "static" and "templates", using local static files instead of ones embedded in binary`)
			}
		}

		os.DirFS(".")
		w, err := web.New(web.Config{
			Logger:     log,
			ListenAddr: c.String("listen-addr"),
		}, webDir)
		if err != nil {
			log.Panicf("error starting web listener: %s", err)
		}
		if len(c.String("pprof-addr")) > 0 {
			log.Infof("listening pprof on %s", c.String("pprof-addr"))
			go func() {
				log.Errorf("failed to start debug listener: %s (ignoring)", http.ListenAndServe(c.String("pprof-addr"), nil))
			}()
		}
		return w.Run()
	}
	// to sort do that
	// sort.Sort(cli.FlagsByName(app.Flags))
	// sort.Sort(cli.CommandsByName(app.Commands))
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

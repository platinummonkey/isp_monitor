package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-yaml/yaml"
	"github.com/platinummonkey/isp-monitor/collectors"
	"github.com/platinummonkey/isp-monitor/config"
	logger "github.com/platinummonkey/isp-monitor/log"
	"github.com/platinummonkey/isp-monitor/reporters"
	"go.uber.org/zap"
)

var options = struct {
	config string
	debug  bool
}{
	config: "~/.isp_monitor.yaml",
}

func init() {
	flag.StringVar(&options.config, "config", options.config, "config file to use")
	flag.BoolVar(&options.debug, "debug", false, "enable debug mode")
}

func main() {
	flag.Parse()

	logger.Initialize(options.debug)
	logger.Get().Info("starting ISP monitor...")

	var err error
	var cfg config.Config
	contents, err := ioutil.ReadFile(options.config)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file") {
			logger.Get().Fatal("no such configuration file exists...")
			logger.Get().Sync()
			os.Exit(1)
		}
		contents = []byte{}
	}
	if len(contents) > 0 {
		err = yaml.Unmarshal(contents, &cfg)
		if err != nil {
			logger.Get().Fatal("invalid configuration file...", zap.Error(err))
			logger.Get().Sync()
			os.Exit(1)
		}
	}

	statReporters := make(map[string]reporters.Interface, 0)
	if len(cfg.Reporters) == 0 {
		// assume log only
		cfg.Reporters = append(cfg.Reporters, config.Section{
			Name: "defaultLog",
			Type: "log",
		})
	}
	for _, c := range cfg.Reporters {
		rep := reporters.CreateReporterFromConfig(c, options.debug)
		if rep != nil {
			statReporters[rep.Name()] = rep
		}
	}

	statCollectors := make(map[string]collectors.Interface, 0)
	if len(cfg.Collectors) == 0 {
		logger.Get().Fatal("no collectors are configured! Please configure collectors!")
		logger.Get().Sync()
		os.Exit(1)
	}

	for _, c := range cfg.Collectors {
		col := collectors.CreateCollectorFromConfig(c, options.debug)
		if col != nil {
			statCollectors[col.Name()] = col
		}
	}

	// start running all collectors
	for _, c := range statCollectors {
		c.Run(statReporters)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// wait for exit
	go func() {
		<-sigs
		logger.Get().Info("exiting...")
		logger.Get().Sync()
		done <- true
	}()
	<-done
	os.Exit(0)
}

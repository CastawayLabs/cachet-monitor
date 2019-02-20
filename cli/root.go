package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"

	cachet "github.com/castawaylabs/cachet-monitor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cmd",
	Short: "cachet-monitor",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		Action(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&cfgFile, "config", "c", "", "config file (default is $(pwd)/config.yml)")
	pf.String("log", "", "log output")
	pf.String("format", "text", "log format [text/json]")
	pf.String("name", "", "machine name")
	pf.Bool("immediate", false, "Tick immediately (by default waits for first defined")
}

func Action(cmd *cobra.Command, args []string) {
	cfg, err := cachet.New(cfgFile)
	if err != nil {
		logrus.Panicf("Unable to start (reading config): %v", err)
	}

	if immediate, err := cmd.Flags().GetBool("immediate"); err == nil && immediate {
		cfg.Immediate = immediate
	}
	if name, err := cmd.Flags().GetString("name"); err == nil && len(name) > 0 {
		cfg.SystemName = name
	}

	logrus.SetOutput(getLogger(cmd))
	if format, err := cmd.Flags().GetString("format"); err == nil && format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	if valid := cfg.Validate(); !valid {
		logrus.Errorf("Invalid configuration")
		os.Exit(1)
	}

	logrus.Debug("Configuration valid")
	logrus.Infof("System: %s", cfg.SystemName)
	// logrus.Infof("API: %s", cfg.API.URL)
	logrus.Infof("Monitors: %d", len(cfg.Monitors))
	logrus.Infof("Backend: %v", strings.Join(cfg.Backend.Describe(), "\n - "))

	logrus.Infof("Pinging backend")
	if err := cfg.Backend.Ping(); err != nil {
		logrus.Errorf("Cannot ping backend!\n%v", err)
		// os.Exit(1)
	}
	logrus.Infof("Ping OK")
	logrus.Warnf("Starting!")

	wg := &sync.WaitGroup{}
	for index, monitor := range cfg.Monitors {
		logrus.Infof("Starting Monitor #%d: ", index)
		logrus.Infof("Features: \n - %v", strings.Join(monitor.Describe(), "\n - "))

		go monitor.Start(monitor.GetTestFunc(), wg, cfg.Backend.Tick, cfg.Immediate)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals

	logrus.Warnf("Abort: Waiting for monitors to finish")
	for _, mon := range cfg.Monitors {
		mon.GetMonitor().Stop()
	}

	wg.Wait()
}

func getLogger(cmd *cobra.Command) *os.File {
	logPath, _ := cmd.Flags().GetString("log")
	if len(logPath) == 0 {
		return os.Stdout
	}

	file, err := os.Create(logPath)
	if err != nil {
		logrus.Errorf("Unable to open file '%v' for logging: \n%v", logPath, err)
		os.Exit(1)
	}

	return file
}

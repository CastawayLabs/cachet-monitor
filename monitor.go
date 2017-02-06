package cachet

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

const DefaultInterval = time.Second * 60
const DefaultTimeout = time.Second
const DefaultTimeFormat = "15:04:05 Jan 2 MST"
const HistorySize = 10

type MonitorInterface interface {
	ClockStart(*CachetMonitor, MonitorInterface, *sync.WaitGroup)
	ClockStop()
	tick(MonitorInterface)
	test() bool

	Validate() []string
	GetMonitor() *AbstractMonitor
	Describe() []string
}

// AbstractMonitor data model
type AbstractMonitor struct {
	Name   string
	Target string

	// (default)http, tcp, dns, icmp
	Type   string
	Strict bool

	Interval time.Duration
	Timeout  time.Duration

	MetricID    int `mapstructure:"metric_id"`
	ComponentID int `mapstructure:"component_id"`

	// Templating stuff
	Template struct {
		Investigating MessageTemplate
		Fixed         MessageTemplate
	}

	// Threshold = percentage
	Threshold float32

	history        []bool
	lastFailReason string
	incident       *Incident
	config         *CachetMonitor

	// Closed when mon.Stop() is called
	stopC chan bool
}

func (mon *AbstractMonitor) Validate() []string {
	errs := []string{}

	if len(mon.Name) == 0 {
		errs = append(errs, "Name is required")
	}

	if mon.Interval < 1 {
		mon.Interval = DefaultInterval
	}
	if mon.Timeout < 1 {
		mon.Timeout = DefaultTimeout
	}

	if mon.Timeout > mon.Interval {
		errs = append(errs, "Timeout greater than interval")
	}

	if mon.ComponentID == 0 && mon.MetricID == 0 {
		errs = append(errs, "component_id & metric_id are unset")
	}

	if mon.Threshold <= 0 {
		mon.Threshold = 100
	}

	if err := mon.Template.Fixed.Compile(); err != nil {
		errs = append(errs, "Could not compile template: "+err.Error())
	}

	return errs
}
func (mon *AbstractMonitor) GetMonitor() *AbstractMonitor {
	return mon
}
func (mon *AbstractMonitor) Describe() []string {
	features := []string{"Type: " + mon.Type}

	if len(mon.Name) > 0 {
		features = append(features, "Name: "+mon.Name)
	}

	return features
}

func (mon *AbstractMonitor) ClockStart(cfg *CachetMonitor, iface MonitorInterface, wg *sync.WaitGroup) {
	wg.Add(1)
	mon.config = cfg
	mon.stopC = make(chan bool)
	if cfg.Immediate {
		mon.tick(iface)
	}

	ticker := time.NewTicker(mon.Interval * time.Second)
	for {
		select {
		case <-ticker.C:
			mon.tick(iface)
		case <-mon.stopC:
			wg.Done()
			return
		}
	}
}

func (mon *AbstractMonitor) ClockStop() {
	select {
	case <-mon.stopC:
		return
	default:
		close(mon.stopC)
	}
}

func (mon *AbstractMonitor) test() bool { return false }

func (mon *AbstractMonitor) tick(iface MonitorInterface) {
	reqStart := getMs()
	up := iface.test()
	lag := getMs() - reqStart

	if len(mon.history) == HistorySize-1 {
		logrus.Warnf("%v is now saturated\n", mon.Name)
	}
	if len(mon.history) >= HistorySize {
		mon.history = mon.history[len(mon.history)-(HistorySize-1):]
	}
	mon.history = append(mon.history, up)
	mon.AnalyseData()

	// report lag
	if mon.MetricID > 0 {
		go mon.config.API.SendMetric(mon.MetricID, lag)
	}
}

// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (monitor *AbstractMonitor) AnalyseData() {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range monitor.history {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(monitor.history))) * 100
	logrus.Printf("%s %.2f%%/%.2f%% down at %v\n", monitor.Name, t, monitor.Threshold, time.Now().UnixNano()/int64(time.Second))

	if len(monitor.history) != HistorySize {
		// not saturated
		return
	}

	if t > monitor.Threshold && monitor.incident == nil {
		monitor.incident = &Incident{
			Name:        monitor.Name + " - " + monitor.config.SystemName,
			ComponentID: monitor.ComponentID,
			Message:     monitor.Name + " check **failed** - " + time.Now().Format(DefaultTimeFormat),
			Notify:      true,
		}

		if len(monitor.lastFailReason) > 0 {
			monitor.incident.Message += "\n\n `" + monitor.lastFailReason + "`"
		}

		// is down, create an incident
		logrus.Printf("%v creating incident. Monitor is down: %v", monitor.Name, monitor.lastFailReason)
		// set investigating status
		monitor.incident.SetInvestigating()
		// create/update incident
		if err := monitor.incident.Send(monitor.config); err != nil {
			logrus.Printf("Error sending incident: %v\n", err)
		}
	} else if t < monitor.Threshold && monitor.incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		logrus.Printf("%v resolved downtime incident", monitor.Name)

		// resolve incident
		monitor.incident.Message = "\n**Resolved** - " + time.Now().Format(DefaultTimeFormat) + "\n\n - - - \n\n" + monitor.incident.Message
		monitor.incident.SetFixed()
		monitor.incident.Send(monitor.config)

		monitor.lastFailReason = ""
		monitor.incident = nil
	}
}

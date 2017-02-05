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
	do() bool
	Validate() []string
	GetMonitor() AbstractMonitor
}

// AbstractMonitor data model
type AbstractMonitor struct {
	Name   string
	Target string

	// (default)http, tcp, dns, icmp
	Type string

	// defaults true
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

func (mon *AbstractMonitor) do() bool {
	return true
}
func (mon *AbstractMonitor) Validate() []string {
	return []string{}
}
func (mon AbstractMonitor) GetMonitor() AbstractMonitor {
	return mon
}

func (mon *AbstractMonitor) Start(cfg *CachetMonitor, wg *sync.WaitGroup) {
	wg.Add(1)
	mon.config = cfg
	mon.stopC = make(chan bool)
	mon.Tick()

	ticker := time.NewTicker(mon.Interval * time.Second)
	for {
		select {
		case <-ticker.C:
			mon.Tick()
		case <-mon.stopC:
			wg.Done()
			return
		}
	}
}

func (monitor *AbstractMonitor) Stop() {
	if monitor.Stopped() {
		return
	}

	close(monitor.stopC)
}

func (monitor *AbstractMonitor) Stopped() bool {
	select {
	case <-monitor.stopC:
		return true
	default:
		return false
	}
}

func (monitor *AbstractMonitor) Tick() {
	reqStart := getMs()
	up := monitor.do()
	lag := getMs() - reqStart

	if len(monitor.history) == HistorySize-1 {
		logrus.Warnf("%v is now saturated\n", monitor.Name)
	}
	if len(monitor.history) >= HistorySize {
		monitor.history = monitor.history[len(monitor.history)-(HistorySize-1):]
	}
	monitor.history = append(monitor.history, up)
	monitor.AnalyseData()

	// report lag
	if up && monitor.MetricID > 0 {
		logrus.Infof("%v", lag)
		// monitor.SendMetric(lag)
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

	if len(monitor.history) != 10 {
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

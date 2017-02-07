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

	// Threshold = percentage / number of down incidents
	Threshold      float32
	ThresholdCount bool `mapstructure:"threshold_count"`

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
		errs = append(errs, "Could not compile \"fixed\" template: "+err.Error())
	}
	if err := mon.Template.Investigating.Compile(); err != nil {
		errs = append(errs, "Could not compile \"investigating\" template: "+err.Error())
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

	histSize := HistorySize
	if mon.ThresholdCount {
		histSize = int(mon.Threshold)
	}

	if len(mon.history) == histSize-1 {
		logrus.Warnf("%v is now saturated\n", mon.Name)
	}
	if len(mon.history) >= histSize {
		mon.history = mon.history[len(mon.history)-(histSize-1):]
	}
	mon.history = append(mon.history, up)
	mon.AnalyseData()

	// report lag
	if mon.MetricID > 0 {
		go mon.config.API.SendMetric(mon.MetricID, lag)
	}
}

// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (mon *AbstractMonitor) AnalyseData() {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range mon.history {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(mon.history))) * 100
	if mon.ThresholdCount {
		logrus.Printf("%s %d/%d down at %v", mon.Name, numDown, int(mon.Threshold), time.Now().Format(mon.config.DateFormat))
	} else {
		logrus.Printf("%s %.2f%%/%.2f%% down at %v", mon.Name, t, mon.Threshold, time.Now().Format(mon.config.DateFormat))
	}

	histSize := HistorySize
	if mon.ThresholdCount {
		histSize = int(mon.Threshold)
	}

	if len(mon.history) != histSize {
		// not saturated
		return
	}

	triggered := (mon.ThresholdCount && numDown == int(mon.Threshold)) || (!mon.ThresholdCount && t > mon.Threshold)

	if triggered && mon.incident == nil {
		// create incident
		tplData := getTemplateData(mon)
		tplData["FailReason"] = mon.lastFailReason

		subject, message := mon.Template.Investigating.Exec(tplData)
		mon.incident = &Incident{
			Name:        subject,
			ComponentID: mon.ComponentID,
			Message:     message,
			Notify:      true,
		}

		// is down, create an incident
		logrus.Warnf("%v: creating incident. Monitor is down: %v", mon.Name, mon.lastFailReason)
		// set investigating status
		mon.incident.SetInvestigating()
		// create/update incident
		if err := mon.incident.Send(mon.config); err != nil {
			logrus.Printf("Error sending incident: %v\n", err)
		}

		return
	}

	// still triggered or no incident
	if triggered || mon.incident == nil {
		return
	}

	logrus.Warnf("Resolving incident")

	// was down, created an incident, its now ok, make it resolved.
	logrus.Printf("%v resolved downtime incident", mon.Name)

	// resolve incident
	mon.incident.Message = "\n**Resolved** - " + time.Now().Format(mon.config.DateFormat) + "\n\n - - - \n\n" + mon.incident.Message
	mon.incident.SetFixed()
	mon.incident.Send(mon.config)

	mon.lastFailReason = ""
	mon.incident = nil
}

package monitors

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const DefaultInterval = time.Second * 60
const DefaultTimeout = time.Second
const HistorySize = 10

type MonitorStatus string

const (
	MonitorStatusUp           = "up"
	MonitorStatusDown         = "down"
	MonitorStatusNotSaturated = "unsaturated"
)

type backendValidateFunc = func(monitor *AbstractMonitor) []string
type MonitorTestFunc func() (up bool, errs []error)
type MonitorTickFunc func(monitor MonitorInterface, status MonitorStatus, errs []error, lag int64)

type MonitorInterface interface {
	Start(MonitorTestFunc, *sync.WaitGroup, MonitorTickFunc, bool)
	Stop()

	tick(MonitorTestFunc) (status MonitorStatus, errors []error, lag int64)
	test() (bool, []error)

	Validate(validate backendValidateFunc) []string
	Describe() []string

	GetMonitor() *AbstractMonitor
	GetTestFunc() MonitorTestFunc
	GetLastStatus() MonitorStatus
	UpdateLastStatus(status MonitorStatus) (old MonitorStatus)
}

// AbstractMonitor data model
type AbstractMonitor struct {
	Name   string
	Target string

	// (default)http / dns
	Type   string
	Strict bool

	Interval time.Duration
	Timeout  time.Duration
	Params   map[string]interface{}

	// Templating stuff
	Template MonitorTemplates

	// Threshold = percentage / number of down incidents
	Threshold      float32
	ThresholdCount bool `mapstructure:"threshold_count"`

	// lag / average(lagHistory) * 100 = percentage above average lag
	// PerformanceThreshold sets the % limit above which this monitor will trigger degraded-performance
	// PerformanceThreshold float32

	history    []bool
	lastStatus MonitorStatus

	// Closed when mon.Stop() is called
	stopC chan bool
}

func (mon *AbstractMonitor) Validate(validate backendValidateFunc) []string {
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

	// get the backend to validate the monitor
	errs = append(errs, validate(mon)...)

	if mon.Threshold <= 0 {
		mon.Threshold = 100
	}

	// if len(mon.Template.Fixed.Message) == 0 || len(mon.Template.Fixed.Subject) == 0 {
	// 	errs = append(errs, "\"fixed\" template empty/missing")
	// }
	// if len(mon.Template.Investigating.Message) == 0 || len(mon.Template.Investigating.Subject) == 0 {
	// 	errs = append(errs, "\"investigating\" template empty/missing")
	// }
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

func (mon *AbstractMonitor) Start(testFunc MonitorTestFunc, wg *sync.WaitGroup, tickFunc MonitorTickFunc, immediate bool) {
	wg.Add(1)

	mon.stopC = make(chan bool)
	if immediate {
		status, errs, lag := mon.tick(testFunc)
		tickFunc(mon, status, errs, lag)
	}

	ticker := time.NewTicker(mon.Interval * time.Second)
	for {
		select {
		case <-ticker.C:
			status, errs, lag := mon.tick(testFunc)
			tickFunc(mon, status, errs, lag)
		case <-mon.stopC:
			wg.Done()
			return
		}
	}
}

func (mon *AbstractMonitor) Stop() {
	select {
	case <-mon.stopC:
		return
	default:
		close(mon.stopC)
	}
}

func (mon *AbstractMonitor) tick(testFunc MonitorTestFunc) (status MonitorStatus, errors []error, lag int64) {
	reqStart := getMs()
	up, errs := testFunc()
	lag = getMs() - reqStart

	histSize := HistorySize
	if mon.ThresholdCount {
		histSize = int(mon.Threshold)
	}

	if len(mon.history) == histSize-1 {
		logrus.WithFields(logrus.Fields{
			"monitor": mon.Name,
		}).Warn("monitor saturated")
	}
	if len(mon.history) >= histSize {
		mon.history = mon.history[len(mon.history)-(histSize-1):]
	}
	mon.history = append(mon.history, up)
	status = mon.GetStatus()
	errors = errs

	return
}

// TODO: test
// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (mon *AbstractMonitor) GetStatus() MonitorStatus {
	numDown := 0
	for _, wasUp := range mon.history {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(mon.history))) * 100
	logFields := logrus.Fields{"monitor": mon.Name}
	// stop reporting time for jsonformatter, it's there by default
	if _, ok := logrus.StandardLogger().Formatter.(*logrus.JSONFormatter); !ok {
		logFields["t"] = time.Now()
	}
	l := logrus.WithFields(logFields)

	symbol := "⚠️"
	if t == 100 {
		symbol = "❌"
	}
	if numDown == 0 {
		l.Printf("👍 up")
	} else if mon.ThresholdCount {
		l.Printf("%v down (%d/%d)", symbol, numDown, int(mon.Threshold))
	} else {
		l.Printf("%v down %.0f%%/%.0f%%", symbol, t, mon.Threshold)
	}

	histSize := HistorySize
	if mon.ThresholdCount {
		histSize = int(mon.Threshold)
	}

	if len(mon.history) != histSize {
		// not saturated
		return MonitorStatusNotSaturated
	}

	var down bool
	if mon.ThresholdCount {
		down = numDown >= int(mon.Threshold)
	} else {
		down = t >= mon.Threshold
	}

	if !down {
		return MonitorStatusUp
	}

	return MonitorStatusDown
}

func (mon *AbstractMonitor) GetTestFunc() MonitorTestFunc {
	return mon.test
}

func (mon *AbstractMonitor) GetLastStatus() MonitorStatus {
	return mon.lastStatus
}

func (mon *AbstractMonitor) UpdateLastStatus(status MonitorStatus) (old MonitorStatus) {
	old = mon.lastStatus
	mon.lastStatus = status

	return
}

func (mon *AbstractMonitor) test() (bool, []error) { return false, nil }

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

package backends

import (
	"net/http"

	"github.com/castawaylabs/cachet-monitor/monitors"
)

type BackendInterface interface {
	Ping() error
	Tick(monitor monitors.MonitorInterface, status monitors.MonitorStatus, errs []error, lag int64)
	SendMetric(monitor monitors.MonitorInterface, lag int64) error
	UpdateMonitor(monitor monitors.MonitorInterface, status, previousStatus monitors.MonitorStatus, errs []error) error
	NewRequest(requestType, url string, reqBody []byte) (*http.Response, interface{}, error)

	Describe() []string
	Validate() []string
	ValidateMonitor(monitor *monitors.AbstractMonitor) []string
}

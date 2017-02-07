package cachet

import (
	"testing"
)

func TestGetMonitorType(t *testing.T) {
	if monType := GetMonitorType(""); monType != "http" {
		t.Error("monitor type `` should default to http")
	}

	if mt := GetMonitorType("HTTP"); mt != "http" {
		t.Error("does not return correct monitor type")
	}
}

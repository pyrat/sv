package winch

import "github.com/CreateView/dawinchi-hw/internal/hwlog"

type HWLog interface {
	AddEvent(event *hwlog.Event)
	Close() error
}

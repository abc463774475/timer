//go:build windows
// +build windows

package timewheel

import (
	"container/list"
	"time"

	nlog "github.com/abc463774475/my_tool/n_log"
)

// NewTimeWheel returns a new TimeWheel.
func NewTimeWheel(interval time.Duration, slotsNum int) *TimeWheel {
	if int(interval)%slotsNum != 0 {
		panic("interval % slotsNum != 0")
	}

	// windows 精度 不能小于 30ms， 所以windows下面最小间隔是 15ms这里必须单独分开！
	tickeTime := interval / time.Duration(slotsNum)
	if tickeTime < 20*time.Millisecond {
		nlog.Erro("windows 精度 不能小于 20ms， 所以windows下面最小间隔是 20ms")
		tickeTime = 20 * time.Millisecond
		interval = tickeTime * time.Duration(slotsNum)
	}

	// nlog.Info("tickeTime:%d", tickeTime)
	tw := &TimeWheel{
		interval:    tickeTime,
		ticker:      time.NewTicker(tickeTime),
		slots:       make([]*list.List, slotsNum),
		slotsNum:    slotsNum,
		currentSlot: 0,
		items:       make(map[int64]*Item),
		quitChan:    make(chan struct{}, 10),
	}

	return tw
}

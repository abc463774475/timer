package timer

import (
	"math"
	"testing"
	"time"

	nlog "github.com/abc463774475/my_tool/n_log"
	"github.com/ouqiang/timewheel"
	_ "github.com/ouqiang/timewheel"
)

// TestTimer is a basic test for timer
func TestTimer(t *testing.T) {
	go func() {
		time.Sleep(1000 * time.Second)
		nlog.Info("sleep 1000 second")
	}()
	tm := time.AfterFunc(1*time.Second, func() {
		nlog.Info("hello world")

		time.Sleep(5 * time.Second)

		nlog.Info("hello world 2")
	})

	time.Sleep(3 * time.Second)
	b := tm.Stop()

	nlog.Info("hello world 3 %v", b)

	if !b {
		<-tm.C
	}
	nlog.Info("hello world 4")

	time.Sleep(10 * time.Second)
}

func TestTimer1(t *testing.T) {
	nlog.Info("hello world 1")
	go func() {
		time.Sleep(1000 * time.Second)
		nlog.Info("sleep 1000 second")
	}()
	tm := time.NewTimer(2 * time.Second)
	nlog.Info("hello world 2")
	if !tm.Stop() {
		nlog.Info("hello world 33333333")
		<-tm.C
	}

	nlog.Info("hello world 3")
}

func TestNewTimer(t *testing.T) {
	f1 := math.Ceil(1.5)
	f2 := math.Floor(1.5)
	f3 := math.Round(1.5)
	f4 := math.Round(1.4)

	nlog.Info("f1 %v f2 %v f3 %v f4 %v", f1, f2, f3, f4)
}

// TestNewtimers
func TestNewtimers(t *testing.T) {
}

// TestTimeWheel  一款很锤子的时间轮
func TestTimeWheel(t *testing.T) {
	nlog.Info("start")
	tw := timewheel.New(1*time.Second, 3600, func(i interface{}) {
		nlog.Info("hello world %v", i)
	})

	tw.Start()

	tw.AddTimer(1*time.Second, 1, 1111111)

	time.Sleep(1000 * time.Second)
}

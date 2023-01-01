package timewheel

import (
	"math/rand"
	"testing"
	"time"

	nlog "github.com/abc463774475/my_tool/n_log"
)

// TestTimeWheel is a test function
func TestTimeWheel(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	nlog.Info("start %v", time.Now())
	tw := NewTimeWheel(1*time.Second, 50)

	tw.Start()

	tw.Add(2*time.Second, -1, func() {
		nlog.Info("3 %v", time.Now())
	})

	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	nlog.Debug("current time %v", time.Now())
	tw.Add(3*time.Second, 3, func() {
		nlog.Debug("1 %v", time.Now())
	})

	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	nlog.Erro("curtime %v", time.Now())
	tw.Add(4*time.Second, 5, func() {
		nlog.Erro("2 %v", time.Now())
	})

	time.Sleep(20000 * time.Second)
}

// TestTimeWheel2 is a test function
func TestTimeWheel2(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	nlog.Info("start %v", time.Now())
	tw := NewTimeWheel(1*time.Second, 50)

	tw.Start()

	// time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	nlog.Debug("current time %v", time.Now())
	tw.Add(1*time.Second, 100, func() {
		nlog.Debug("1 %v", time.Now())
	})

	time.Sleep(400 * time.Second)
}

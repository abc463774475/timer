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
	}, nil)

	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	nlog.Debug("current time %v", time.Now())
	tw.Add(3*time.Second, 3, func() {
		nlog.Debug("1 %v", time.Now())
	}, nil)

	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	nlog.Erro("curtime %v", time.Now())
	tw.Add(4*time.Second, 5, func() {
		nlog.Erro("2 %v", time.Now())
	}, nil)

	time.Sleep(20000 * time.Second)
}

// TestTimeWheel2 is a test function
func TestTimeWheel2(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	nlog.Info("start %v", time.Now())
	tw := NewTimeWheel(1*time.Second, 50)

	tw.Start()

	ts := NewItems(tw, func(item *Item) {
		nlog.Info("this is items %v", item)
		item.GetCallback()()
	})

	ts.Add(2*time.Second, -1, func() {
		nlog.Info("3 %v", time.Now())
	})

	t1 := ts.Add(3*time.Second, 3, func() {
		nlog.Debug("1 %v", time.Now())
	})

	time.Sleep(7 * time.Second)
	nlog.Info("clear")
	ts.Clear()

	t1.ResetDuration(1*time.Second, 2)

	time.Sleep(400 * time.Second)
}

// TestTimeWheel3 is a test function
func TestTimeWheel3(t *testing.T) {
	nlog.InitLog(nlog.WithCompressType(nlog.Quick))
	nlog.Info("start %v", time.Now())
	tw := NewTimeWheel(1*time.Second, 100)

	tw.Start()

	// time.Sleep(269 * time.Millisecond)

	startTime := time.Now()

	count := 0
	var it *Item
	it = tw.Add(3*time.Second, 3, func() {
		nlog.Erro("%v", time.Now().Sub(startTime))
		count++
		if count > 1 {
			it.Stop()
			//_, bret := it.ResetDuration(10*time.Second, 5)
			//nlog.Erro("reset %v", bret)
		}
	}, nil)

	time.Sleep(100 * time.Second)
}

// TestTimeWheel4 is a test function
func TestTimeWheel4(t *testing.T) {
	nlog.Info("start %v", time.Now())
	t2 := time.AfterFunc(5*time.Second, func() {
		nlog.Info("t2 %v", time.Now())
	})

	time.Sleep(5 * time.Second)
	ret := t2.Reset(2 * time.Second)
	nlog.Info("finish %v %v", time.Now(), ret)
	time.Sleep(100 * time.Second)
}

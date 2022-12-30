package timer

import (
	"testing"
	"time"

	nlog "github.com/abc463774475/my_tool/n_log"
	"github.com/rfyiamcool/go-timewheel"
	_ "github.com/rfyiamcool/go-timewheel"
)

// TestTimeWheel1  一款很锤子的时间轮
func TestTimeWheel1(t *testing.T) {
	nlog.Info("start %v", time.Now())
	tw, err := timewheel.NewTimeWheel(10*time.Millisecond,
		10)
	if err != nil {
		nlog.Erro("err %v", err)
		return
	}

	tw.Start()
	defer tw.Stop()

	tw.Add(10*time.Millisecond, func() {
		nlog.Info("hello world %v", time.Now())
	})

	t1 := tw.AfterFunc(10*time.Millisecond, func() {
		nlog.Info("hello world 2 %v", time.Now())
	})

	t1.Reset(10 * time.Second)

	time.Sleep(100 * time.Second)
}

// TestTimeWheel2  一款很锤子的时间轮
func TestTimeWheel2(t *testing.T) {
	start := time.Now()
	t1 := time.AfterFunc(10*time.Second, func() {
		nlog.Info("hello world 2 %v", time.Now().Sub(start))
	})

	t1.Reset(-1)

	time.Sleep(10 * time.Second)
}

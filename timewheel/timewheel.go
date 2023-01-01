package timewheel

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	nlog "github.com/abc463774475/my_tool/n_log"
)

var timerIndex int64

type Item struct {
	// id
	id int64
	// 延迟时间
	delay time.Duration
	// 回调函数
	callback func()
	// counts -1 无限循环 0 一次 1 两次 2 三次
	counts int
	// round , 第几轮
	round int
	// slot , 第几个槽
	slot int
	// element
	element *list.Element
}

type TimeWheel struct {
	// interval is the time duration between two ticks.
	interval time.Duration
	// ticker is the ticker that ticks every interval.
	ticker *time.Ticker
	// slots
	slots []*list.List
	// slotsNum
	slotsNum int
	// currentSlot
	currentSlot int
	// allSlots
	allSlots int

	l sync.Mutex

	items map[int64]*Item

	curRound int

	quitChan chan struct{}
}

// Start starts the TimeWheel.
func (tw *TimeWheel) Start() {
	go func() {
		defer func() {
			nlog.Info("TimeWheel Stop")
		}()

		for {
			select {
			case <-tw.ticker.C:
				tw.advance()
			case <-tw.quitChan:
				return
			}
		}
	}()
}

// Add adds a new item to the TimeWheel.
func (tw *TimeWheel) Add(delay time.Duration, counts int, callback func()) *Item {
	if counts < 0 {
		counts = -1
	} else if counts == 0 {
		counts = 1
	}

	item := &Item{
		id:       atomic.AddInt64(&timerIndex, 1),
		delay:    delay,
		callback: callback,
		counts:   counts,
	}

	tw.l.Lock()
	tw.items[item.id] = item
	tw.add(item, true)

	tw.l.Unlock()

	return item
}

// add adds an item to the TimeWheel.
func (tw *TimeWheel) add(item *Item, isNextSlot bool) {
	// 计算延迟时间
	round := int(item.delay / (time.Duration(tw.slotsNum) * tw.interval))

	nextSlot := tw.currentSlot
	if isNextSlot {
		nextSlot--
	}
	// 计算槽位
	slot := (nextSlot + int(item.delay/tw.interval)) % tw.slotsNum

	if slot < 0 {
		// 按常理来说，这里不会出现小于0的情况，但是为了防止出现bug，这里做了处理
		panic("slot < 0")
	}

	// nlog.Erro("round:%d,slot:%d curslots %v", round, slot, nextSlot)

	item.round = round
	item.slot = slot

	if tw.slots[slot] == nil {
		tw.slots[slot] = list.New()
	}

	item.element = tw.slots[slot].PushBack(item)
}

// Remove removes an item from the TimeWheel.
func (tw *TimeWheel) Remove(id int64) {
	tw.l.Lock()
	defer tw.l.Unlock()

	item, ok := tw.items[id]
	if !ok {
		return
	}

	tw.remove(item)
}

// remove removes an item from the TimeWheel.
func (tw *TimeWheel) remove(item *Item) {
	tw.slots[item.slot].Remove(item.element)
	delete(tw.items, item.id)
}

// Stop stops the TimeWheel.
func (tw *TimeWheel) Stop() {
	tw.ticker.Stop()

	tw.l.Lock()
	defer tw.l.Unlock()

	for _, item := range tw.items {
		tw.remove(item)
	}

	tw.items = make(map[int64]*Item)
}

// Reset resets the TimeWheel.
func (tw *TimeWheel) Reset(id int64, duration time.Duration) {
	tw.l.Lock()
	defer tw.l.Unlock()

	item, ok := tw.items[id]
	if !ok {
		return
	}

	item.delay = duration
	tw.remove(item)
	tw.add(item, false)
}

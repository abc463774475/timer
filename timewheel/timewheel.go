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

// advance advances the TimeWheel by one tick.
func (tw *TimeWheel) advance() {
	tw.l.Lock()
	defer tw.l.Unlock()

	tw.currentSlot %= tw.slotsNum

	if tw.currentSlot%tw.slotsNum == 0 {
		tw.curRound++
	}

	slots := tw.slots[tw.currentSlot]
	if slots == nil {
		tw.currentSlot++

		return
	}

	tw.currentSlot++

	addItems := []*Item{}
	removeItems := []*Item{}
	// 遍历槽位
	for e := slots.Front(); e != nil; e = e.Next() {
		item := e.Value.(*Item)
		item.round--

		// nlog.Info("item.id  %v item.round:%d", item.id, item.round)
		if item.round == 0 {
			item.callback()
			if item.counts == 0 {
			} else if item.counts <= -1 {
				addItems = append(addItems, item)
			} else {
				item.counts--
				if item.counts > 0 {
					addItems = append(addItems, item)
				}
			}
			removeItems = append(removeItems, item)
		}
	}

	for _, item := range removeItems {
		tw.remove(item)
	}

	for _, item := range addItems {
		tw.add(item)
	}
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
	tw.add(item)

	tw.l.Unlock()

	return item
}

// add adds an item to the TimeWheel.
func (tw *TimeWheel) add(item *Item) {
	// 计算延迟时间
	round := int(item.delay/(time.Duration(tw.slotsNum)*tw.interval)) + 1

	// 计算槽位
	slot := (tw.currentSlot + int(item.delay/tw.interval)) % tw.slotsNum

	// nlog.Erro("round:%d,slot:%d curslots %v", round, slot, tw.currentSlot)

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

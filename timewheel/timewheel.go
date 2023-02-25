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
	// round int
	// slot , 第几个槽
	slot int
	// element
	element *list.Element
	// Items
	Items *Items
	// tw
	tw *TimeWheel
	// frame
	frame int
}

// GetCallback gets the callback function of the item.
func (it *Item) GetCallback() func() {
	return it.callback
}

// Reset
func (it *Item) ResetDuration(duration time.Duration, times int) {
	if it.Items != nil {
		it.Items.tw.l.Lock()
		defer it.Items.tw.l.Unlock()

		it.Items.tw.remove(it)
		it.delay = duration
		it.counts = times
		it.Items.tw.add(it)
	} else {
		it.delay = duration
		it.counts = times

		it.tw.l.Lock()
		defer it.tw.l.Unlock()

		it.tw.remove(it)
		it.tw.add(it)
	}
}

type Items struct {
	Items    map[int64]*Item
	itemFunc func(item *Item)

	tw *TimeWheel
	l  sync.RWMutex
}

// NewItems
func NewItems(tw *TimeWheel, itemFunc func(item *Item)) *Items {
	return &Items{
		Items:    make(map[int64]*Item),
		itemFunc: itemFunc,
		tw:       tw,
	}
}

// Add
func (items *Items) Add(delay time.Duration, counts int, callback func()) *Item {
	item := items.tw.Add(delay, counts, callback, items)
	items.l.Lock()
	items.Items[item.id] = item
	// nlog.Info("Add item %v", item.id)
	items.l.Unlock()

	return item
}

// Remove
func (items *Items) Remove(id int64) {
	items.l.Lock()
	defer items.l.Unlock()

	item, ok := items.Items[id]
	if !ok {
		return
	}

	items.tw.Remove(item.id)
	delete(items.Items, item.id)
}

// Clear clears the Items
func (i *Items) Clear() {
	i.l.Lock()
	defer i.l.Unlock()

	for _, item := range i.Items {
		i.tw.Remove(item.id)
	}

	i.Items = make(map[int64]*Item)
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

	// addItems
	addItems []*Item
	// removeItems
	removeItems []*Item

	// frame
	frame int
}

// Start starts the TimeWheel.
func (tw *TimeWheel) Start() {
	go func() {
		defer func() {
			nlog.Info("TimeWheel Stop")
		}()

		startTime := time.Now()

		frame := 0
		for {
			select {
			case <-tw.ticker.C:
				curTime := time.Now()
				subTime := curTime.Sub(startTime)
				curFrame := int(subTime / tw.interval)
				if curFrame > 0 && curFrame > frame {
					// nlog.Info("curFrame %v frame %v", curFrame, frame)
					for i := 0; i < curFrame-frame; i++ {
						tw.advance()
						frame++
					}

					tw.advanceAfter()
				}
			case <-tw.quitChan:
				return
			}
		}
	}()
}

// Add adds a new item to the TimeWheel.
func (tw *TimeWheel) Add(delay time.Duration,
	counts int, callback func(), items *Items,
) *Item {
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
		Items:    items,
		tw:       tw,
	}

	tw.l.Lock()
	tw.items[item.id] = item
	tw.add(item)

	// nlog.Erro("Add item %v", item.id)
	tw.l.Unlock()

	return item
}

// add adds an item to the TimeWheel.
func (tw *TimeWheel) add(item *Item) {
	// 计算延迟时间
	//round := int(item.delay / (time.Duration(tw.slotsNum) * tw.interval))
	//
	//nextSlot := tw.currentSlot
	// 计算槽位
	//slot := (nextSlot + int(item.delay/tw.interval)) % tw.slotsNum
	//
	//if slot < 0 {
	//	// 按常理来说，这里不会出现小于0的情况，但是为了防止出现bug，这里做了处理
	//	panic("slot < 0")
	//}
	//r := int(item.delay / (time.Duration(tw.slotsNum) * tw.interval))
	//
	needFrame := int(item.delay/tw.interval) + tw.frame
	item.frame = needFrame
	slot := needFrame % tw.slotsNum
	// item.round = round
	item.slot = slot

	nlog.Info("add item %v  slot %v", item.id, slot)

	if tw.slots[slot] == nil {
		tw.slots[slot] = list.New()
	}

	item.element = tw.slots[slot].PushBack(item)

	if _, ok := tw.items[item.id]; !ok {
		tw.items[item.id] = item
	}
}

// Remove removes an item from the TimeWheel.
func (tw *TimeWheel) Remove(id int64) {
	tw.l.Lock()
	defer tw.l.Unlock()

	item, ok := tw.items[id]
	if !ok {
		nlog.Debug("Remove item not found %v", id)
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
//func (tw *TimeWheel) Reset(id int64, duration time.Duration) {
//	tw.l.Lock()
//	defer tw.l.Unlock()
//
//	item, ok := tw.items[id]
//	if !ok {
//		return
//	}
//
//	item.delay = duration
//	tw.remove(item)
//	tw.add(item, false)
//}

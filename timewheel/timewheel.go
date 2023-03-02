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
	// totalCounts -1 无限循环 0 一次 1 两次 2 三次
	totalCounts int
	// currentCounts
	currentCounts int
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
	// createTime
	createTime int64
}

// GetCallback gets the callback function of the item.
func (it *Item) GetCallback() func() {
	return it.callback
}

// ResetDuration
// 本来不想给这个函数的，可能会导致时间轮的混乱，但是我想到了一个办法，就是在重置的时候，先把这个item从时间轮中删除，然后再重新添加到时间轮中
// 这样就可以保证时间轮的正确性
// 现在限定这个函数的功能
// 返回值为true，表示重置成功, false 表示重置失败, 重置失败的原因是，这个item已经被删除了 这个返回值很重要
// 成功调用之后 , 尤其是reset 函数，，所以 必须判定 是否存在，这个item的延迟时间会被重置为duration，这个item的执行次数会被重置为times。但是开始时间会以当前时间为准
// 还是慎用这个函数吧，最好应用层自己删除+添加，因为这个函数会导致时间轮的混乱
// 多携程的时候，这个函数会有问题
func (it *Item) ResetDuration(duration time.Duration, times int) (*Item, bool) {
	it.tw.l.Lock()
	if _, ok := it.tw.items[it.id]; !ok {
		it.tw.l.Unlock()
		nlog.Erro("ResetDuration failed, item has been deleted, id %v", it.id)
		return nil, false
	}
	it.tw.remove(it)
	it.tw.l.Unlock()
	return it.tw.Add(duration, times, it.callback, it.Items), true
}

// Stop
func (it *Item) Stop() {
	it.tw.l.Lock()
	defer it.tw.l.Unlock()

	if _, ok := it.tw.items[it.id]; !ok {
		return
	}

	it.tw.remove(it)
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
	// startTime
	startTime time.Time
	once      sync.Once

	l sync.Mutex

	items map[int64]*Item

	curRound int

	quitChan chan struct{}

	// frame
	frame int
}

// Start starts the TimeWheel.
func (tw *TimeWheel) Start() {
	tw.startTime = time.Now()

	go func() {
		defer func() {
			nlog.Info("TimeWheel Stop")
		}()

		frame := 0
		for {
			select {
			case <-tw.ticker.C:
				curTime := time.Now()
				subTime := curTime.Sub(tw.startTime)
				curFrame := int(subTime / tw.interval)
				if curFrame > 0 && curFrame > frame {
					for i := 0; i < curFrame-frame; i++ {
						tw.advance()
						frame++
					}
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
		id:          atomic.AddInt64(&timerIndex, 1),
		delay:       delay,
		callback:    callback,
		totalCounts: counts,
		Items:       items,
		tw:          tw,
		createTime:  time.Now().UnixNano(),
	}

	tw.add(item)
	return item
}

// add adds an item to the TimeWheel.
func (tw *TimeWheel) add(item *Item) {
	// todo 后面如果轮训量还是太大，可以考虑分层查看，，
	// 比如 slot =0 的 有 1层 2层 3层 等，每次只轮训当前层，已有层全部舍掉
	nextTime := item.createTime + ((int64(item.currentCounts) + 1) * item.delay.Nanoseconds())
	sub := nextTime - tw.startTime.UnixNano()

	needFrame := int(sub / int64(tw.interval))

	item.frame = needFrame
	slot := needFrame % tw.slotsNum
	item.slot = slot

	if tw.slots[slot] == nil {
		tw.slots[slot] = list.New()
	}

	item.element = tw.slots[slot].PushBack(item)

	if _, ok := tw.items[item.id]; !ok {
		tw.items[item.id] = item
	}
}

// Remove removes an item from the TimeWheel.
// 这是错的，，callback 里面调用的时候，会导致死锁
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
	if item.element == nil {
		return
	}

	tw.slots[item.slot].Remove(item.element)
	delete(tw.items, item.id)
	item.element = nil
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

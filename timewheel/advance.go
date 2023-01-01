package timewheel

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

	addItems := []*Item{}
	removeItems := []*Item{}
	// 遍历槽位
	for e := slots.Front(); e != nil; e = e.Next() {
		item := e.Value.(*Item)
		item.round--

		// nlog.Debug("item.id  %v item.round:%d  cur slots %v", item.id, item.round, tw.currentSlot)
		if item.round <= 0 {
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
		tw.add(item, false)
	}

	tw.currentSlot++
}

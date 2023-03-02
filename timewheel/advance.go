package timewheel

// advance advances the TimeWheel by one tick..
func (tw *TimeWheel) advance() {
	tw.l.Lock()
	defer func() {
		tw.frame++
		tw.l.Unlock()
	}()

	tw.currentSlot %= tw.slotsNum

	slots := tw.slots[tw.currentSlot]
	tw.currentSlot++
	if slots == nil || slots.Len() == 0 {
		return
	}

	updateItems := make([]*Item, 0)
	removeItems := make([]*Item, 0)
	addItems := make([]*Item, 0)

	// 遍历槽位
	for e := slots.Front(); e != nil; e = e.Next() {
		item := e.Value.(*Item)

		if tw.frame != item.frame {
			continue
		}

		updateItems = append(updateItems, item)
		removeItems = append(removeItems, item)

		if item.currentCounts < item.totalCounts-1 {
			addItems = append(addItems, item)
		}
	}

	for _, item := range updateItems {
		if item.Items != nil {
			item.Items.itemFunc(item)
		} else {
			go item.callback()
		}
	}

	for _, item := range removeItems {
		if item.Items != nil {
			item.Items.Remove(item.id)
		} else {
			tw.remove(item)
		}
	}

	for _, item := range addItems {
		item.currentCounts++
		tw.add(item)
	}
}

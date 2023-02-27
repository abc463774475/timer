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
	if slots == nil {
		return
	}

	// 遍历槽位
	for e := slots.Front(); e != nil; e = e.Next() {
		item := e.Value.(*Item)

		if tw.frame == item.frame {
			// 如果item的上层控制不为nil，则交给上层处理
			if item.Items != nil {
				item.Items.itemFunc(item)
			} else {
				go item.callback()
			}

			if item.totalCounts == 0 {
			} else if item.totalCounts <= -1 {
				tw.addItems = append(tw.addItems, item)
			} else {
				item.totalCounts--
				if item.totalCounts > 0 {
					tw.addItems = append(tw.addItems, item)
				}
			}
			tw.removeItems = append(tw.removeItems, item)
		}
	}
}

// advanceAfter
func (tw *TimeWheel) advanceAfter() {
	for _, item := range tw.removeItems {
		tw.remove(item)
	}

	for _, item := range tw.addItems {
		item.currentCounts++
		tw.add(item)
	}

	tw.removeItems = tw.removeItems[:0]
	tw.addItems = tw.addItems[:0]
}

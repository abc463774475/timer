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
		// tFrame := item.round*tw.slotsNum + tw.currentSlot
		// item.round--

		// nlog.Debug("item.id  %v item.round:%d  cur slots %v", item.id, item.round, tw.currentSlot)
		if tw.frame == item.frame {
			// 如果item的上层控制不为nil，则交给上层处理
			if item.Items != nil {
				item.Items.itemFunc(item)
				// item.callback()
			} else {
				go item.callback()
			}

			if item.counts == 0 {
			} else if item.counts <= -1 {
				tw.addItems = append(tw.addItems, item)
			} else {
				item.counts--
				if item.counts > 0 {
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
		tw.add(item)
	}

	tw.removeItems = tw.removeItems[:0]
	tw.addItems = tw.addItems[:0]
}

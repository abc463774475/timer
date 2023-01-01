package timewheel

// advance advances the TimeWheel by one tick..
// todo 应该用上次的时间差来计算，而不是用固定的时间间隔。暂时不会出太大的问题，但是如果时间间隔很大，就会出问题了。后面再改
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
			// 无解，多携程的东西！不是我想要的！战斗系统不能这么写！
			// item.callback()

			// 如果item的上层控制不为nil，则交给上层处理
			if item.Items != nil {
				item.Items.itemFunc(item)
				// item.callback()
			} else {
				go item.callback()
			}

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

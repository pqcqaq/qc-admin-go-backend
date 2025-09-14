package utils

// DiffUint64Slices 比较两个 uint64 切片，返回需要删除和添加的元素
// oldSlice: 旧的切片
// newSlice: 新的切片
// 返回值:
//   - toRemove: 在 oldSlice 中存在但在 newSlice 中不存在的元素
//   - toAdd: 在 newSlice 中存在但在 oldSlice 中不存在的元素
func DiffUint64Slices(oldSlice, newSlice []uint64) (toRemove, toAdd []uint64) {
	// 创建 map 用于快速查找
	oldMap := make(map[uint64]bool)
	newMap := make(map[uint64]bool)

	// 将旧切片的元素放入 map
	for _, id := range oldSlice {
		oldMap[id] = true
	}

	// 将新切片的元素放入 map
	for _, id := range newSlice {
		newMap[id] = true
	}

	// 找出需要删除的元素（在旧切片中但不在新切片中）
	for id := range oldMap {
		if !newMap[id] {
			toRemove = append(toRemove, id)
		}
	}

	// 找出需要添加的元素（在新切片中但不在旧切片中）
	for id := range newMap {
		if !oldMap[id] {
			toAdd = append(toAdd, id)
		}
	}

	return toRemove, toAdd
}

// DiffUint64SlicesOrdered 有序版本，保持元素在原切片中的顺序
func DiffUint64SlicesOrdered(oldSlice, newSlice []uint64) (toRemove, toAdd []uint64) {
	// 创建 map 用于快速查找
	oldMap := make(map[uint64]bool)
	newMap := make(map[uint64]bool)

	// 将切片元素放入 map
	for _, id := range oldSlice {
		oldMap[id] = true
	}
	for _, id := range newSlice {
		newMap[id] = true
	}

	// 按原顺序找出需要删除的元素
	for _, id := range oldSlice {
		if !newMap[id] {
			toRemove = append(toRemove, id)
		}
	}

	// 按原顺序找出需要添加的元素
	for _, id := range newSlice {
		if !oldMap[id] {
			toAdd = append(toAdd, id)
		}
	}

	return toRemove, toAdd
}

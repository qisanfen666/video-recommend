package internal

import (
	"container/heap"
	"fmt"
	"sort"
	"time"
)

type Item struct {
	VideoID  int64
	Title    string
	Heat     float64
	Category string
}

type MinHeap []Item

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Heat < h[j].Heat }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(Item))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

var (
	hotHeap *MinHeap
	hotMap  map[int64]*Item
)

func Init() {
	fmt.Println("[i1] 程序启动:", time.Now().Format("15:04:05"))
	hotHeap = &MinHeap{}
	heap.Init(hotHeap)
	hotMap = make(map[int64]*Item, 500000)

	for _, video := range VideoIndex {
		update(video)
	}
}

func update(video Video) {
	//已经在堆中，更新热度
	if item, ok := hotMap[video.ID]; ok {
		item.Heat = video.Heat
		heap.Fix(hotHeap, findIndex(item))
		return
	}

	//堆未满，直接加入
	if hotHeap.Len() < 20 {
		item := Item{
			VideoID:  video.ID,
			Title:    video.Title,
			Heat:     video.Heat,
			Category: video.Category,
		}
		heap.Push(hotHeap, item)
		hotMap[video.ID] = &item
		return
	}

	// 超过第20名，直接替换
	if video.Heat > (*hotHeap)[0].Heat {
		old := heap.Pop(hotHeap).(Item)
		delete(hotMap, old.VideoID)

		item := Item{
			VideoID:  video.ID,
			Title:    video.Title,
			Heat:     video.Heat,
			Category: video.Category,
		}
		heap.Push(hotHeap, item)
		hotMap[video.ID] = &item
	}
}

func findIndex(target *Item) int {
	for i, item := range *hotHeap {
		if item.VideoID == target.VideoID {
			return i
		}
	}
	return -1
}

func GetTop20() []Item {
	result := make([]Item, hotHeap.Len())
	copy(result, *hotHeap)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Heat > result[j].Heat
	})

	return result
}
